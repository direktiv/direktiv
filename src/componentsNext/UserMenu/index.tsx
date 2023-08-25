import {
  CurlyBraces,
  LogOut,
  Moon,
  Settings2,
  Slack,
  Sun,
  Terminal,
  User,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { useTheme, useThemeActions } from "~/util/store/theme";

import Avatar from "~/design/Avatar";
import Button from "~/design/Button";
import { RxChevronDown } from "react-icons/rx";
import { twMergeClsx } from "~/util/helpers";
import { useApiActions } from "~/util/store/apiKey";
import useApiKeyHandling from "~/hooksNext/useApiKeyHandling";
import { useTranslation } from "react-i18next";

interface UserMenuProps {
  className?: string;
}

const UserMenu: React.FC<UserMenuProps> = ({ className }) => {
  const { isKeyRequired } = useApiKeyHandling();
  const { setTheme } = useThemeActions();
  const theme = useTheme();
  const { t } = useTranslation();
  const { setApiKey: storeApiKey } = useApiActions();

  const logout = () => {
    storeApiKey(null);
  };

  return (
    <div className={twMergeClsx("flex space-x-2", className)}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          {isKeyRequired ? (
            <Button
              variant="ghost"
              className="items-center px-1"
              role="button"
              icon
              data-testid="dropdown-trg-user-menu"
            >
              <Avatar>
                <User />
              </Avatar>
              <RxChevronDown />
            </Button>
          ) : (
            <Button variant="ghost" icon data-testid="dropdown-trg-user-menu">
              <Settings2 />
              <RxChevronDown />
            </Button>
          )}
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          {isKeyRequired && (
            <>
              <DropdownMenuLabel>
                {t("components.userMenu.loggedIn")}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={logout} className="cursor-pointer">
                <LogOut className="mr-2 h-4 w-4" />
                <span>{t("components.userMenu.logout")}</span>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
            </>
          )}

          <DropdownMenuLabel>
            {t("components.userMenu.appearance")}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            data-testid="dropdown-item-switch-theme"
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
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
          <DropdownMenuItem>
            <Terminal className="mr-2 h-4 w-4" />
            {t("components.userMenu.showApiCommands")}
          </DropdownMenuItem>
          <DropdownMenuItem>
            <CurlyBraces className="mr-2 h-4 w-4" />
            {t("components.userMenu.openJQPlayground")}
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <a
              href="https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g"
              target="_blank"
              rel="noopener noreferrer"
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
