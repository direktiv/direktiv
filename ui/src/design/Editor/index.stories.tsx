import {
  CSSSample,
  HtmlSample,
  JsonSample,
  PlaintextSample,
  ShellSample,
} from "./languageSamples";
import { Card } from "../Card";
import Editor from "./index";
import type { Meta } from "@storybook/react-vite";
import { hello } from "~/pages/namespace/Explorer/Tree/components/modals/CreateNew/Workflow/templates";

export default {
  title: "Components/Editor",
} satisfies Meta<typeof Editor>;

const tsExample = hello.data;

export const Default = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={tsExample} />
    </div>
  </div>
);

export const Small = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="size-[500px]">
      <Editor value={tsExample} />
    </div>
  </div>
);
export const Darkmode = () => (
  <div className="flex flex-col gap-y-3 bg-black p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={tsExample} theme="dark" />
    </div>
  </div>
);

export const WithCardAnd100Height = () => (
  <div className="flex h-[97vh] min-h-full flex-col gap-y-3 bg-black">
    <div>This Story is not aware of light and dark mode.</div>
    <Card className="grow p-4">
      <Editor value={tsExample} theme="dark" />
    </Card>
  </div>
);

export const HtmlEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={HtmlSample} language="html" />
    </div>
  </div>
);
export const DarkHtmlEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={HtmlSample} language="html" theme="dark" />
    </div>
  </div>
);

export const CSSEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={CSSSample} language="css" />
    </div>
  </div>
);
export const DarkCSSEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={CSSSample} language="css" theme="dark" />
    </div>
  </div>
);

export const JsonEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={JsonSample} language="json" />
    </div>
  </div>
);
export const DarkJsonEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={JsonSample} language="json" theme="dark" />
    </div>
  </div>
);

export const ShellEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={ShellSample} language="shell" />
    </div>
  </div>
);
export const DarkShellEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={ShellSample} language="shell" theme="dark" />
    </div>
  </div>
);

export const PlaintextEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={PlaintextSample} language="plaintext" />
    </div>
  </div>
);
export const DarkPlaintextEditor = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={PlaintextSample} language="plaintext" theme="dark" />
    </div>
  </div>
);
