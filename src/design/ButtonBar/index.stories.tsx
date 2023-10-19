import {
  Bug,
  ChevronDown,
  GitBranchIcon,
  HelpCircle,
  SearchIcon,
  Undo,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../Dropdown";
import type { Meta, StoryObj } from "@storybook/react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../Tooltip";
import Button from "../Button";
import { ButtonBar } from "./index";
import { Toggle } from "../Toggle";

const meta = {
  title: "Components/ButtonBar",
  component: ButtonBar,
} satisfies Meta<typeof ButtonBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <ButtonBar {...args}>
      <Button>Start</Button>
      <Button>Mid-1</Button>
      <Button>Mid-2</Button>
      <Button>End</Button>
    </ButtonBar>
  ),
};

export const GitButtonBar = () => (
  <div className="flex gap-2">
    <DropdownMenu>
      <ButtonBar>
        <Button variant="outline">
          <GitBranchIcon /> Review
        </Button>
        <DropdownMenuTrigger asChild>
          <Button variant="outline">
            <ChevronDown />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem>
            <Undo className="mr-2 h-4 w-4" />
            Revert to Previous
          </DropdownMenuItem>
        </DropdownMenuContent>
      </ButtonBar>
    </DropdownMenu>
  </div>
);
export const ButtonSizes = () => (
  <div className="flex flex-col gap-4">
    <ButtonBar>
      <Button size="lg">Start</Button>
      <Button size="lg">Mid-1</Button>
      <Button size="lg">Mid-2</Button>
      <Button size="lg">
        <GitBranchIcon /> Review
      </Button>
      <Button size="lg">
        <SearchIcon /> Search
      </Button>
    </ButtonBar>
    <ButtonBar>
      <Button>Start</Button>
      <Button>Mid-1</Button>
      <Button>Mid-2</Button>
      <Button>
        <GitBranchIcon /> Review
      </Button>
      <Button>
        <SearchIcon /> Search
      </Button>
    </ButtonBar>
    <ButtonBar>
      <Button size="sm">Start</Button>
      <Button size="sm">Mid-1</Button>
      <Button size="sm">Mid-2</Button>
      <Button size="sm">
        <GitBranchIcon /> Review
      </Button>
      <Button size="sm">
        <SearchIcon /> Search
      </Button>
    </ButtonBar>
  </div>
);

export const AsChildButtons = () => (
  <div className="flex flex-col gap-4">
    <p>No Click, Hover effects on child components except Anchor</p>
    <ButtonBar>
      <Button asChild size="lg">
        <label>Label</label>
      </Button>
      <Button asChild size="lg">
        <a href="/">Link</a>
      </Button>
      <Button size="lg">
        <GitBranchIcon /> Review
      </Button>
      <Button size="lg">
        <SearchIcon /> Search
      </Button>
      <Button asChild size="lg">
        <span>Span Tag</span>
      </Button>
    </ButtonBar>
  </div>
);

export const ToolbarWithToolTips = () => (
  <div className="flex flex-col space-y-3">
    <div>
      Please note the extra div in the between the TooltipTrigger and Toggle.
    </div>
    <ButtonBar>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            {/* 
unfortunately this div is required. TooltipTrigger must be used with asChild, 
to avoid having a button inside a button, which is semantically invalid and also
causes design issues with the ButtonBar (double borders). And withtout the extra
div, the asChild would merge the TooltipTrigger and Toggle into one button with 
shared state. The tooltip and and the toggle both need the data-state state and
the toggles state will get lost and it will never show as pressed.

potential solutions are discussed here by the radix-ui team:

https://github.com/radix-ui/primitives/discussions/560
TLDR; It could technically solved, but all state attributes would need to be 
namespaced which would have a DX impact that is not worth it for now.
             */}

            <div>
              <Toggle aria-label="Toggle italic">
                <HelpCircle />
              </Toggle>
            </div>
          </TooltipTrigger>
          <TooltipContent>Hi 👋 from Toggle</TooltipContent>
        </Tooltip>
        <Toggle aria-label="Toggle italic">
          <GitBranchIcon />
        </Toggle>
        <Button variant="outline">
          <SearchIcon />
        </Button>
        <Tooltip>
          <TooltipTrigger asChild>
            <div>
              <Button variant="outline" aria-label="Toggle italic">
                <Bug /> Button with tooltip
              </Button>
            </div>
          </TooltipTrigger>
          <TooltipContent>Hi 👋 from Button</TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </ButtonBar>
  </div>
);
