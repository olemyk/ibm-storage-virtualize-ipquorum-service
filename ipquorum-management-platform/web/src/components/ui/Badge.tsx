import React from 'react';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'secondary' | 'destructive';
  size?: 'sm' | 'md' | 'lg';
  dot?: boolean;
}

const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  ({ children, variant = 'default', size = 'md', dot = false, className = '', ...props }, ref) => {
    const baseStyles = 'inline-flex items-center rounded-full border font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2';

    const variantStyles = {
      default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
      success: 'border-transparent bg-green-100 text-green-800 hover:bg-green-200',
      warning: 'border-transparent bg-yellow-100 text-yellow-800 hover:bg-yellow-200',
      danger: 'border-transparent bg-red-100 text-red-800 hover:bg-red-200',
      info: 'border-transparent bg-blue-100 text-blue-800 hover:bg-blue-200',
      secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
      destructive: 'border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80',
    };

    const sizeStyles = {
      sm: 'px-2 py-0.5 text-xs',
      md: 'px-2.5 py-0.5 text-sm',
      lg: 'px-3 py-1 text-base',
    };

    const dotColors = {
      default: 'bg-primary-foreground',
      success: 'bg-green-600',
      warning: 'bg-yellow-600',
      danger: 'bg-red-600',
      info: 'bg-blue-600',
      secondary: 'bg-secondary-foreground',
      destructive: 'bg-destructive-foreground',
    };

    const combinedClassName = `${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`;

    return (
      <span ref={ref} className={combinedClassName} {...props}>
        {dot && (
          <span
            className={`mr-1.5 h-2 w-2 rounded-full ${dotColors[variant]}`}
            aria-hidden="true"
          />
        )}
        {children}
      </span>
    );
  }
);

Badge.displayName = 'Badge';

export default Badge;


