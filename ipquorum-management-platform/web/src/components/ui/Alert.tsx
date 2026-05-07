import React from 'react';
import { AlertCircle, CheckCircle, Info, XCircle, X } from 'lucide-react';

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'info' | 'success' | 'warning' | 'danger' | 'destructive';
  title?: string;
  dismissible?: boolean;
  onDismiss?: () => void;
}

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(
  (
    {
      children,
      variant = 'info',
      title,
      dismissible = false,
      onDismiss,
      className = '',
      ...props
    },
    ref
  ) => {
    const baseStyles = 'relative w-full rounded-lg border p-4';

    const variantStyles = {
      info: 'bg-blue-50 text-blue-900 border-blue-200 [&>svg]:text-blue-600',
      success: 'bg-green-50 text-green-900 border-green-200 [&>svg]:text-green-600',
      warning: 'bg-yellow-50 text-yellow-900 border-yellow-200 [&>svg]:text-yellow-600',
      danger: 'bg-red-50 text-red-900 border-red-200 [&>svg]:text-red-600',
      destructive: 'border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive',
    };

    const icons = {
      info: Info,
      success: CheckCircle,
      warning: AlertCircle,
      danger: XCircle,
      destructive: XCircle,
    };

    const Icon = icons[variant];
    const combinedClassName = `${baseStyles} ${variantStyles[variant]} ${className}`;

    return (
      <div ref={ref} className={combinedClassName} role="alert" {...props}>
        <div className="flex">
          <div className="flex-shrink-0">
            <Icon className="h-5 w-5" aria-hidden="true" />
          </div>
          <div className="ml-3 flex-1">
            {title && (
              <h5 className="mb-1 font-medium leading-none tracking-tight">{title}</h5>
            )}
            <div className="text-sm [&_p]:leading-relaxed">{children}</div>
          </div>
          {dismissible && onDismiss && (
            <div className="ml-auto pl-3">
              <button
                type="button"
                onClick={onDismiss}
                className="inline-flex rounded-md p-1.5 hover:bg-black/5 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-transparent"
                aria-label="Dismiss"
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </button>
            </div>
          )}
        </div>
      </div>
    );
  }
);

Alert.displayName = 'Alert';

export default Alert;


