export { default as Button } from "./Button.vue"

export type ButtonVariants = {
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link' | 'icon' | null
  size?: 'default' | 'sm' | 'lg' | 'icon' | null
}

export function buttonVariants(options: ButtonVariants = {}) {
  const variant = options.variant || 'default'
  const size = options.size || 'default'
  const base = "relative inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-xl text-sm font-semibold transition-[background-color,border-color,color,box-shadow,transform,opacity] duration-[var(--motion-normal)] ease-out disabled:pointer-events-none disabled:opacity-60 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:border-[var(--brand-400)] focus-visible:ring-4 focus-visible:ring-brand-200/40 aria-invalid:ring-red-200/50 aria-invalid:border-red-400 active:scale-[0.98]"
  const variants = {
    default: "border border-[var(--brand-600)] bg-[var(--brand-600)] text-white shadow-none hover:border-[var(--brand-700)] hover:bg-[var(--brand-700)] hover:-translate-y-px focus-visible:bg-[var(--brand-700)]",
    destructive: "border border-red-600 bg-red-600 text-white shadow-none hover:border-red-700 hover:bg-red-700 hover:-translate-y-px focus-visible:bg-red-700",
    outline: "border border-[var(--dora-border)] bg-[var(--dora-container-bg)] text-foreground shadow-none hover:border-[var(--brand-200)] hover:bg-[var(--brand-50)] hover:text-[var(--brand-600)] hover:-translate-y-px",
    secondary: "border border-[var(--dora-border)] bg-[var(--admin-surface-muted)] text-foreground shadow-none hover:border-[var(--brand-200)] hover:bg-[var(--brand-50)] hover:text-[var(--brand-600)] hover:-translate-y-px",
    ghost: "border border-transparent bg-transparent text-foreground/80 shadow-none hover:border-[var(--brand-100)] hover:bg-[var(--brand-50)] hover:text-[var(--brand-600)]",
    link: "border border-transparent bg-transparent text-[var(--brand-600)] shadow-none underline-offset-4 hover:text-[var(--brand-700)] hover:underline active:scale-100",
    icon: "border border-transparent bg-transparent text-[var(--admin-text-muted)] shadow-none hover:border-[var(--brand-100)] hover:bg-[var(--brand-50)] hover:text-[var(--brand-600)] focus-visible:bg-[var(--brand-50)]",
  }
  const sizes = {
    default: "h-9 px-4 py-2 has-[>svg]:px-3",
    sm: "h-8 gap-1.5 px-3 has-[>svg]:px-2.5",
    lg: "h-10 px-6 has-[>svg]:px-4",
    icon: "size-9",
  }
  return [base, variants[variant], sizes[size]].join(' ')
}
