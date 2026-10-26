import './AlertBadge.css';

// ============================================
// Tipos
// ============================================

export type AlertBadgeVariant = 'critical' | 'warning' | 'info' | 'neutral';

interface AlertBadgeProps {
  /** Número a mostrar */
  count: number;

  /** Color/variante del badge */
  variant?: AlertBadgeVariant;

  /** Texto opcional después del número */
  label?: string;

  /** Máximo a mostrar antes de usar "99+" */
  max?: number;

  /** Si es true, no muestra el badge cuando count es 0 */
  hideWhenZero?: boolean;

  /** Si es true, muestra animación de pulso */
  pulse?: boolean;

  /** Callback al hacer click */
  onClick?: () => void;

  /** Tamaño del badge */
  size?: 'sm' | 'md' | 'lg';

  /** Texto para accesibilidad */
  ariaLabel?: string;
}

// ============================================
// Componente
// ============================================

/**
 * Badge con contador de alertas.
 *
 * @example
 * ```tsx
 * <AlertBadge count={5} variant="critical" onClick={openAlerts} />
 * ```
 */
export function AlertBadge({
  count,
  variant = 'neutral',
  label,
  max = 99,
  hideWhenZero = true,
  pulse = false,
  onClick,
  size = 'md',
  ariaLabel,
}: AlertBadgeProps) {
  // Ocultar si es 0 (y hideWhenZero)
  if (hideWhenZero && count === 0) {
    return null;
  }

  // Formatear el número
  const displayCount = count > max ? `${max}+` : String(count);

  // ¿Es clickeable?
  const isClickable = Boolean(onClick);

  // Texto de accesibilidad
  const accessibilityLabel =
    ariaLabel ||
    (label
      ? `${count} ${label}`
      : `${count} ${count === 1 ? 'alerta' : 'alertas'}`);

  // Clases CSS
  const className = [
    'alert-badge',
    `alert-badge--${variant}`,
    `alert-badge--${size}`,
    isClickable && 'alert-badge--clickable',
    pulse && 'alert-badge--pulse',
  ]
    .filter(Boolean)
    .join(' ');

  // Render como button si es clickeable
  if (isClickable) {
    return (
      <button
        type="button"
        className={className}
        onClick={onClick}
        aria-label={accessibilityLabel}
      >
        <span className="alert-badge__dot" aria-hidden="true" />
        <span className="alert-badge__count">{displayCount}</span>
        {label && <span className="alert-badge__label">{label}</span>}
      </button>
    );
  }

  // Render como span si no es clickeable
  return (
    <span className={className} aria-label={accessibilityLabel}>
      <span className="alert-badge__dot" aria-hidden="true" />
      <span className="alert-badge__count">{displayCount}</span>
      {label && <span className="alert-badge__label">{label}</span>}
    </span>
  );
}