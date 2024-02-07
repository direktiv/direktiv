const LogoElement = ({
  iconOnly,
  theme = "light",
}: {
  iconOnly?: boolean;
  theme?: "light" | "dark";
}): JSX.Element => {
  const path = iconOnly
    ? `/assets/logo/icon-${theme}.svg`
    : `/assets/logo/logo-${theme}.svg`;
  return <img src={path} alt="" />;
};

const Logo = ({
  className,
  iconOnly,
  theme,
}: {
  className?: string;
  iconOnly?: boolean;
  theme?: "light" | "dark";
}): JSX.Element => (
  <div className={className}>
    <div className="flex h-[32px] max-w-[159px]">
      <LogoElement iconOnly={iconOnly} theme={theme} />
    </div>
  </div>
);

export default Logo;
