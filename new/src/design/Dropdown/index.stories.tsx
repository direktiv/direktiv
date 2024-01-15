import {
  ChevronDown,
  Cloud,
  CreditCard,
  Github,
  Keyboard,
  LifeBuoy,
  Loader2,
  LogOut,
  Mail,
  MessageSquare,
  Plus,
  PlusCircle,
  Settings,
  User,
  UserPlus,
  Users,
} from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../Dialog";

import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuPortal,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "./index";
import Button from "../Button";
import { DialogClose } from "@radix-ui/react-dialog";
import type { Meta } from "@storybook/react";
import { useState } from "react";

const meta = {
  title: "Components/Dropdown",
  component: DropdownMenu,
} satisfies Meta<typeof DropdownMenu>;

export default meta;

export const LoadingState = () => (
  <div className="flex justify-center">
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button>
          Open Dropdown <ChevronDown />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56">
        <DropdownMenuLabel>My Account</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem>
            <User className="mr-2 h-4 w-4" />
            <span>Profile</span>
            <DropdownMenuShortcut>⇧⌘P</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem>
            <CreditCard className="mr-2 h-4 w-4" />
            <span>Billing</span>
            <DropdownMenuShortcut>⌘B</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem>
            <Settings className="mr-2 h-4 w-4" />
            <span>Settings</span>
            <DropdownMenuShortcut>⌘S</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem>
            <Keyboard className="mr-2 h-4 w-4" />
            <span>Keyboard shortcuts</span>
            <DropdownMenuShortcut>⌘K</DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem>
            <Users className="mr-2 h-4 w-4" />
            <span>Team</span>
          </DropdownMenuItem>
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>
              <UserPlus className="mr-2 h-4 w-4" />
              <span>Invite users</span>
            </DropdownMenuSubTrigger>
            <DropdownMenuPortal>
              <DropdownMenuSubContent>
                <DropdownMenuItem>
                  <Mail className="mr-2 h-4 w-4" />
                  <span>Email</span>
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <MessageSquare className="mr-2 h-4 w-4" />
                  <span>Message</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <PlusCircle className="mr-2 h-4 w-4" />
                  <span>More...</span>
                </DropdownMenuItem>
              </DropdownMenuSubContent>
            </DropdownMenuPortal>
          </DropdownMenuSub>
          <DropdownMenuItem>
            <Plus className="mr-2 h-4 w-4" />
            <span>New Team</span>
            <DropdownMenuShortcut>⌘+T</DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <Github className="mr-2 h-4 w-4" />
          <span>GitHub</span>
        </DropdownMenuItem>
        <DropdownMenuItem>
          <LifeBuoy className="mr-2 h-4 w-4" />
          <span>Support</span>
        </DropdownMenuItem>
        <DropdownMenuItem disabled>
          <Cloud className="mr-2 h-4 w-4" />
          <span>API</span>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem disabled>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          loading...
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <LogOut className="mr-2 h-4 w-4" />
          <span>Log out</span>
          <DropdownMenuShortcut>⇧⌘Q</DropdownMenuShortcut>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
);
export const WithRadioButtons = () => {
  const [position, setPosition] = useState("one");

  return (
    <div className="flex justify-center">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button>
            Radio Buttons <ChevronDown />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          <DropdownMenuLabel>Choose Option</DropdownMenuLabel>
          <DropdownMenuRadioGroup value={position} onValueChange={setPosition}>
            <DropdownMenuRadioItem value="one">
              Option One
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="two">
              Option Two
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="three">
              Option Three
            </DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};
export const WithCheckboxes = () => {
  const [featureOne, setFeatureOne] = useState(true);
  const [featureTwo, setFeatureTwo] = useState(false);
  const [featureThree, setFeatureThree] = useState(false);

  return (
    <div className="flex justify-center">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button>
            Checkboxes <ChevronDown />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56">
          <DropdownMenuLabel>Choose Feature</DropdownMenuLabel>
          <DropdownMenuCheckboxItem
            checked={featureOne}
            onCheckedChange={setFeatureOne}
          >
            Feature One
          </DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem
            checked={featureTwo}
            onCheckedChange={setFeatureTwo}
            disabled
          >
            Feature Two
          </DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem
            checked={featureThree}
            onCheckedChange={setFeatureThree}
          >
            Feature Three
          </DropdownMenuCheckboxItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};

export const MenuItemInDialogTrigger = () => {
  const [dialog, setDialog] = useState<string>("");

  return (
    <div className="flex justify-center">
      <Dialog>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button>
              Open Dropdown <ChevronDown />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56">
            <DropdownMenuLabel>My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DialogTrigger
                className="w-full"
                onClick={() => {
                  setDialog("profile");
                }}
              >
                <DropdownMenuItem>
                  <User className="mr-2 h-4 w-4" />
                  <span>Profile</span>
                  <DropdownMenuShortcut>⇧⌘P</DropdownMenuShortcut>
                </DropdownMenuItem>
              </DialogTrigger>
              <DialogTrigger
                className="w-full"
                onClick={() => {
                  setDialog("billing");
                }}
              >
                <DropdownMenuItem>
                  <CreditCard className="mr-2 h-4 w-4" />
                  <span>Billing</span>
                  <DropdownMenuShortcut>⌘B</DropdownMenuShortcut>
                </DropdownMenuItem>
              </DialogTrigger>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
        <DialogContent>
          {dialog === "profile" && (
            <>
              <DialogHeader>
                <DialogTitle>
                  <Settings />
                  Profile Dialog Title
                </DialogTitle>
                <DialogDescription>
                  This is a description of the profile dialog.
                </DialogDescription>
              </DialogHeader>
              Profile content goes here
              <DialogFooter>
                <DialogClose asChild>
                  <Button variant="ghost">Cancel</Button>
                </DialogClose>
                <Button>Submit</Button>
              </DialogFooter>
            </>
          )}
          {dialog === "billing" && (
            <>
              <DialogHeader>
                <DialogTitle>
                  <Settings />
                  Dialog Billing
                </DialogTitle>
                <DialogDescription>
                  This is a description of the billing dialog.
                </DialogDescription>
              </DialogHeader>
              Billing content goes here
              <DialogFooter>
                <DialogClose asChild>
                  <Button variant="ghost">Cancel</Button>
                </DialogClose>
                <Button>Submit</Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};
