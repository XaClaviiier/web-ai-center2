import type { ComponentProps } from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../lib/utils';

// Varian dipetakan ke kelas `.tombol` prototype agar visual tetap sama.
const buttonVariants = cva('tombol', {
  variants: {
    variant: {
      utama: 'tombol--utama',
      kedua: 'tombol--kedua',
      bahaya: 'tombol--bahaya',
      terang: 'tombol--terang',
      'garis-terang': 'tombol--garis-terang',
    },
    size: { default: '', kecil: 'tombol--kecil' },
  },
  defaultVariants: { variant: 'utama', size: 'default' },
});

export function Button({ className, variant, size, asChild = false, type, ...props }: ComponentProps<'button'> & VariantProps<typeof buttonVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : 'button';
  return <Comp className={cn(buttonVariants({ variant, size }), className)} type={asChild ? type : (type ?? 'button')} {...props} />;
}

export { buttonVariants };
