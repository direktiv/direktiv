import Button from "./index";
import { VscAccount } from "react-icons/vsc";

export default {
  title: "Design System/Button",
  component: Button,
};

export const ButtonSizes = () => (
  <div className="flex flex-wrap gap-5">
    <Button size="xs">XS Button</Button>
    <Button size="sm">SM Button</Button>
    <Button>Normal Button</Button>
    <Button size="lg">lg Button</Button>
  </div>
);

export const ButtonColors = () => (
  <div className="flex flex-wrap gap-5">
    <Button>Default</Button>
    <Button color="primary">Primary</Button>
    <Button color="secondary">Secondary</Button>
    <Button color="accent">Accent</Button>
    <Button color="ghost">Ghost</Button>
    <Button color="link">Link</Button>
  </div>
);

export const ActiveColors = () => (
  <div className="flex flex-wrap gap-5">
    <Button active>Default</Button>
    <Button active color="primary">
      Primary
    </Button>
    <Button active color="secondary">
      Secondary
    </Button>
    <Button active color="accent">
      Accent
    </Button>
    <Button active color="ghost">
      Ghost
    </Button>
    <Button active color="link">
      Link
    </Button>
  </div>
);

export const StateColors = () => (
  <div className="flex flex-wrap gap-5">
    <Button state="info">Info</Button>
    <Button state="success">Success</Button>
    <Button state="warning">Warning</Button>
    <Button state="error">Error</Button>
  </div>
);

export const Outline = () => (
  <div className="flex flex-wrap gap-5">
    <Button outline>Default</Button>
    <Button outline color="primary">
      Primary
    </Button>
    <Button outline color="secondary">
      Secondary
    </Button>
    <Button outline color="accent">
      Accent
    </Button>
    <Button outline color="ghost">
      Ghost
    </Button>
    <Button outline color="link">
      Link
    </Button>
    <Button outline state="info">
      Info
    </Button>
    <Button outline state="success">
      Success
    </Button>
    <Button outline state="warning">
      Warning
    </Button>
    <Button outline state="error">
      Error
    </Button>
  </div>
);

export const Loading = () => (
  <div className="flex flex-wrap gap-5">
    <Button outline loading>
      Loading
    </Button>
  </div>
);

export const WithIcon = () => (
  <div className="flex flex-wrap gap-5">
    <Button color="primary" className="gap-5">
      Loading <VscAccount />
    </Button>
  </div>
);
