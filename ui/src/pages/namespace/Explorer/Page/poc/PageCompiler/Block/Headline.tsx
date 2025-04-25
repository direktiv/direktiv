import { HeadlineType } from "../../schema/blocks/headline";
import { TemplateString } from "./utils/TemplateString";

type HeadlineProps = {
  blockProps: HeadlineType;
};

export const Headline = ({ blockProps }: HeadlineProps) => (
  <h1 className="text-xl">
    <TemplateString value={blockProps.label} />
  </h1>
);
