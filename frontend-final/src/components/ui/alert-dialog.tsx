import * as AlertDialogPrimitive from '@radix-ui/react-alert-dialog';
import { Button } from './button';

type Variant = 'utama' | 'bahaya';

/** Dialog konfirmasi (setara `konfirmasi()` prototype). `onAnswer(true)` bila dikonfirmasi. */
export function ConfirmDialog({ open, title, text, confirmLabel, variant = 'utama', cancelLabel = 'Batal', onAnswer }: {
  open: boolean; title: string; text: string; confirmLabel: string; variant?: Variant; cancelLabel?: string | null; onAnswer: (yes: boolean) => void;
}) {
  return (
    <AlertDialogPrimitive.Root open={open} onOpenChange={(next) => { if (!next) onAnswer(false); }}>
      <AlertDialogPrimitive.Portal>
        <AlertDialogPrimitive.Overlay className="fixed inset-0 z-50 bg-[var(--warna-overlay)]" />
        <AlertDialogPrimitive.Content className="modal fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 bg-white">
          <div className="modal__isi">
            <AlertDialogPrimitive.Title style={{ fontSize: 'var(--judul-3)' }}>{title}</AlertDialogPrimitive.Title>
            <AlertDialogPrimitive.Description>{text}</AlertDialogPrimitive.Description>
            <div className="modal__aksi">
              {cancelLabel && <Button variant="kedua" onClick={() => onAnswer(false)}>{cancelLabel}</Button>}
              <Button variant={variant} onClick={() => onAnswer(true)}>{confirmLabel}</Button>
            </div>
          </div>
        </AlertDialogPrimitive.Content>
      </AlertDialogPrimitive.Portal>
    </AlertDialogPrimitive.Root>
  );
}
