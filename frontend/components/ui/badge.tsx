import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-all duration-200 hover:shadow-md",
  {
    variants: {
      variant: {
        default: "border-transparent bg-primary text-primary-foreground hover:shadow-primary/20",
        secondary: "border-transparent bg-secondary text-secondary-foreground hover:shadow-secondary/20",
        destructive: "border-transparent bg-destructive text-destructive-foreground hover:shadow-destructive/20",
        outline: "text-foreground border-zinc-700 hover:border-zinc-600",
        success: "border-transparent bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30",
        warning: "border-transparent bg-amber-500/20 text-amber-400 hover:bg-amber-500/30",
      },
    },
    defaultVariants: { variant: "default" },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { Badge, badgeVariants };
