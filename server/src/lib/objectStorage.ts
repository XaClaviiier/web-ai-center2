import {
  CreateBucketCommand,
  DeleteObjectCommand,
  GetObjectCommand,
  HeadBucketCommand,
  ListObjectsV2Command,
  PutObjectCommand,
  S3Client,
} from '@aws-sdk/client-s3';
import {
  namaFileAman,
  namaFileUnik,
  tipeKontenGambar,
  type ImageStorage,
  type StoredImage,
} from './upload';

export interface RustFsStorageConfig {
  endpoint: string;
  accessKey: string | null;
  secretKey: string | null;
  bucket: string;
}

function errorDetails(error: unknown): { name?: string; status?: number } {
  if (error === null || typeof error !== 'object') return {};
  const name = 'name' in error && typeof error.name === 'string' ? error.name : undefined;
  const metadata = '$metadata' in error && typeof error.$metadata === 'object' && error.$metadata !== null
    ? error.$metadata
    : null;
  const status =
    metadata !== null &&
    'httpStatusCode' in metadata &&
    typeof metadata.httpStatusCode === 'number'
      ? metadata.httpStatusCode
      : undefined;
  return { name, status };
}

function isBucketMissing(error: unknown): boolean {
  const details = errorDetails(error);
  return details.name === 'NotFound' || details.name === 'NoSuchBucket' || details.status === 404;
}

async function bodyToBytes(body: unknown): Promise<Uint8Array> {
  const stream = body as {
    transformToByteArray?: () => Promise<Uint8Array>;
  };
  if (typeof stream.transformToByteArray === 'function') return stream.transformToByteArray();

  const chunks: Uint8Array[] = [];
  let size = 0;
  for await (const chunk of body as AsyncIterable<Uint8Array>) {
    const bytes = chunk instanceof Uint8Array ? chunk : new Uint8Array(chunk);
    chunks.push(bytes);
    size += bytes.byteLength;
  }
  const result = new Uint8Array(size);
  let offset = 0;
  for (const chunk of chunks) {
    result.set(chunk, offset);
    offset += chunk.byteLength;
  }
  return result;
}

async function ensureBucket(client: S3Client, bucket: string): Promise<void> {
  try {
    await client.send(new HeadBucketCommand({ Bucket: bucket }));
  } catch (error) {
    if (!isBucketMissing(error)) throw error;
    await client.send(new CreateBucketCommand({ Bucket: bucket }));
  }
}

export async function createRustFsImageStorage(config: RustFsStorageConfig): Promise<ImageStorage> {
  if (config.accessKey === null || config.secretKey === null) {
    throw new Error('RUSTFS_ACCESS_KEY dan RUSTFS_SECRET_KEY wajib diisi');
  }

  const client = new S3Client({
    endpoint: config.endpoint,
    // RustFS uses S3-compatible signing; region is an adapter detail.
    region: 'us-east-1',
    credentials: {
      accessKeyId: config.accessKey,
      secretAccessKey: config.secretKey,
    },
    forcePathStyle: true,
  });
  await ensureBucket(client, config.bucket);

  return {
    async save(file, ext) {
      const nama = namaFileUnik(ext);
      await client.send(
        new PutObjectCommand({
          Bucket: config.bucket,
          Key: nama,
          Body: new Uint8Array(await file.arrayBuffer()),
          ContentType: tipeKontenGambar(ext),
          CacheControl: 'public, max-age=31536000, immutable',
        }),
      );
      return nama;
    },

    async get(nama): Promise<StoredImage | null> {
      if (namaFileAman(nama) === null) return null;
      try {
        const result = await client.send(new GetObjectCommand({ Bucket: config.bucket, Key: nama }));
        if (result.Body === undefined) return null;
        return {
          body: await bodyToBytes(result.Body),
        };
      } catch (error) {
        if (isBucketMissing(error) || errorDetails(error).name === 'NoSuchKey') return null;
        throw error;
      }
    },

    async list(): Promise<string[]> {
      const names: string[] = [];
      let continuationToken: string | undefined;
      do {
        const result = await client.send(
          new ListObjectsV2Command({
            Bucket: config.bucket,
            ContinuationToken: continuationToken,
          }),
        );
        for (const object of result.Contents ?? []) {
          if (object.Key !== undefined && namaFileAman(object.Key) !== null) names.push(object.Key);
        }
        continuationToken = result.IsTruncated ? result.NextContinuationToken : undefined;
      } while (continuationToken !== undefined);
      return names.sort();
    },

    async delete(nama): Promise<void> {
      if (namaFileAman(nama) === null) throw new Error(`Nama file tidak aman: ${nama}`);
      await client.send(new DeleteObjectCommand({ Bucket: config.bucket, Key: nama }));
    },
  };
}
