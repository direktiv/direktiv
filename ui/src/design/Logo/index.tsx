import { useTheme } from "~/util/store/theme";

const LogoElement = ({ iconOnly }: { iconOnly?: boolean }): JSX.Element => {
  const themeSetting = useTheme();
  const theme = themeSetting === "dark" ? "dark" : "light";

  const path = iconOnly
    ? `/assets/logo/icon-${theme}.svg`
    : `/assets/logo/logo-${theme}.svg`;
  return <img src={path} alt="" />;
};

const Logo = ({
  className,
  iconOnly,
}: {
  className?: string;
  iconOnly?: boolean;
}): JSX.Element => (
  <div className={className}>
    <div className="flex h-[32px] max-w-[159px]">
      <LogoElement iconOnly={iconOnly} />
    </div>
  </div>
);

export default Logo;
