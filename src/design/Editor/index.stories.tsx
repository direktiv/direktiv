import Editor from "./index";

import type { Meta } from "@storybook/react";

export default {
  title: "Components/Editor",
} satisfies Meta<typeof Editor>;

export const Default = () => <Editor className="h-24" />;
