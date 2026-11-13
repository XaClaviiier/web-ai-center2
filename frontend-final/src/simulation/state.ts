import { useSyncExternalStore } from 'react';
export const SIMULASI: { keadaan: 'normal' | 'loading' | 'kosong' | 'error' | '429'; jedaMs: number; paksa401: boolean; paksa409: boolean; heartbeatGagal: boolean; kunciLain: boolean } = { keadaan: 'normal', jedaMs: 300, paksa401: false, paksa409: false, heartbeatGagal: false, kunciLain: true };
let version = 0;
const listeners = new Set<() => void>();
if (typeof document !== 'undefined') document.addEventListener('simulasi:ubah', () => { version++; listeners.forEach(listener => listener()); });
const subscribe = (listener: () => void) => { listeners.add(listener); return () => { listeners.delete(listener); }; };
export function useSimulationVersion(): number { return useSyncExternalStore(subscribe, () => version, () => 0); }
