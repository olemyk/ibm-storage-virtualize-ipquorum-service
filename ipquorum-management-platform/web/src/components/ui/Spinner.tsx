import React from 'react';
import { Loader2 } from 'lucide-react';

export interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  variant?: 'default' | 'secondary' | 'white';
  label?: string;
  centered?: boolean;
}

const Spinner = React.forwardRef<HTMLDivElement, SpinnerProps>(
  (
    {
      size = 'md',
      variant = 'default',
      label,
      centered = false,
      className = '',
      ...props
    },
    ref
  ) => {
    const sizeStyles = {
      sm: 'h-4 w-4',
      md: 'h-6 w-6',
      lg: 'h-8 w-8',
      xl: 'h-12 w-12',
    };

    const variantStyles = {
      default: 'text-primary',
      secondary: 'text-muted-foreground',
      white: 'text-primary-foreground',
    };

    const containerClassName = centered
      ? 'flex items-center justify-center'
      : 'inline-flex items-center';

    return (
      <div
        ref={ref}
        className={`${containerClassName} ${className}`}
        role="status"
        aria-live="polite"
        {...props}
      >
        <Loader2
          className={`animate-spin ${sizeStyles[size]} ${variantStyles[variant]}`}
          aria-hidden="true"
        />
        {label && (
          <span className="ml-2 text-sm text-muted-foreground">{label}</span>
        )}
        <span className="sr-only">{label || 'Loading...'}</span>
      </div>
    );
  }
);

Spinner.displayName = 'Spinner';

export default Spinner;


