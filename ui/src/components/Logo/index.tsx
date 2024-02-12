import LogoDesignComponent from "~/design/Logo";
import { useTheme } from "~/util/store/theme";

const Logo = ({
  className,
  iconOnly,
}: {
  className?: string;
  iconOnly?: boolean;
}): JSX.Element => {
  const theme = useTheme();
  const logoTheme = theme === "dark" ? "dark" : "light";

  return (
    <LogoDesignComponent
      className={className}
      theme={logoTheme}
      iconOnly={iconOnly}
    />
  );
};

export default Logo;
