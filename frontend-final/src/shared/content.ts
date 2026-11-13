import seed from './seed.json';
export const konten = seed.konten;
const KEY_MEDIA = /^(berita|dokumentasi|testimoni|tim)\/[0-9a-f-]{36}\.(jpg|png|webp)$/;
export const image = (key?: string | null): string => key?.startsWith('blob:') ? key : key && KEY_MEDIA.test(key) ? `/media/${key}` : `/assets/img/${key ?? ''}`;
