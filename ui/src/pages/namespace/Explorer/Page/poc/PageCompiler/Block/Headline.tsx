import { HeadlineType } from "../../schema/blocks/headline";
import { TemplateString } from "./utils/TemplateString";

type HeadlineProps = {
  blockProps: HeadlineType;
};

export const Headline = ({ blockProps: { label } }: HeadlineProps) => (
  <h1 className="text-xl">
    <TemplateString value={label} />
  </h1>
);
