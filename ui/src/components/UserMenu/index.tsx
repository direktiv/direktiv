import { BookOpen, Moon, Slack, Sun } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { useTheme, useThemeActions } from "~/util/store/theme";

import MenuButton from "./MenuButton";
import UserInfo from "./UserInfo";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

interface UserMenuProps {
  className?: string;
}

const UserMenu: React.FC<UserMenuProps> = ({ className }) => {
  const { setTheme } = useThemeActions();
  const theme = useTheme();
  const { t } = useTranslation();

  return (
    <div className={twMergeClsx("flex space-x-2", className)}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <MenuButton />
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          <UserInfo />
          <DropdownMenuLabel>
            {t("components.userMenu.appearance")}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            data-testid="dropdown-item-switch-theme"
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="cursor-pointer"
          >
            {theme === "dark" ? (
              <>
                <Sun className="mr-2 h-4 w-4" />
                {t("components.userMenu.switchToLight")}
              </>
            ) : (
              <>
                <Moon className="mr-2 h-4 w-4" />
                {t("components.userMenu.switchToDark")}
              </>
            )}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuLabel>{t("components.userMenu.help")}</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem asChild>
            <a
              href="https://docs.direktiv.io/"
              target="_blank"
              rel="noopener noreferrer"
              className="cursor-pointer"
            >
              <BookOpen className="mr-2 h-4 w-4" />
              {t("components.userMenu.docs")}
            </a>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem asChild>
            <a
              href="https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g"
              target="_blank"
              rel="noopener noreferrer"
              className="cursor-pointer"
            >
              <Slack className="mr-2 h-4 w-4" />
              {t("components.userMenu.supportChannelOnSlack")}
            </a>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};

export default UserMenu;
