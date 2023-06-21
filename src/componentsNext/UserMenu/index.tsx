import {
  CurlyBraces,
  LogOut,
  Moon,
  Settings2,
  Slack,
  Sun,
  Terminal,
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
import { NavigationLink } from "~/design/NavigationLink";
import { RxChevronDown } from "react-icons/rx";
import clsx from "clsx";

interface UserMenuProps {
  hasAccount?: boolean;
  username?: string;
  className?: string;
}

const UserMenu: React.FC<UserMenuProps> = ({
  hasAccount = true,
  username = "admin",
  className,
}) => {
  const { setTheme } = useThemeActions();
  const theme = useTheme();
  return (
    <div className={clsx("flex space-x-2", className)}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          {hasAccount ? (
            <Button
              variant="ghost"
              className="items-center px-1"
              role="button"
              icon
              data-testid="dropdown-trg-user-menu"
            >
              <Avatar>{username?.slice(0, 2)}</Avatar>
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
          {hasAccount && (
            <>
              <DropdownMenuLabel>You are logged in as admin</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <LogOut className="mr-2 h-4 w-4" />
                <span>Logout</span>
              </DropdownMenuItem>
            </>
          )}
          <DropdownMenuLabel>Appearance</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            data-testid="dropdown-item-switch-mode"
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
          >
            {theme === "dark" ? (
              <>
                <Sun className="mr-2 h-4 w-4" />
                switch to Light mode
              </>
            ) : (
              <>
                <Moon className="mr-2 h-4 w-4" />
                switch to Dark mode
              </>
            )}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuLabel>Help</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem>
            <Terminal className="mr-2 h-4 w-4" /> Show API Commands
          </DropdownMenuItem>
          <DropdownMenuItem>
            <CurlyBraces className="mr-2 h-4 w-4" /> Open JQ Playground
          </DropdownMenuItem>
          <DropdownMenuItem>
            <NavigationLink
              className="p-0"
              href="https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g"
            >
              <Slack className="mr-2 h-4 w-4" /> Support Channel on Slack
            </NavigationLink>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};

export default UserMenu;
