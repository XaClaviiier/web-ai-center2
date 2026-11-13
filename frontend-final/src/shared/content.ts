import seed from './seed.json';
export const konten = seed.konten;
export const image = (key?: string | null): string => key?.startsWith('blob:') ? key : `/assets/img/${key ?? ''}`;
