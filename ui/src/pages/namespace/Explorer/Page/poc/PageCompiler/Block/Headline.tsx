import { HeadlineType } from "../../schema/blocks/headline";
import { TemplateString } from "./utils/TemplateString";
import { twMergeClsx } from "~/util/helpers";

type HeadlineProps = {
  blockProps: HeadlineType;
};

export const Headline = ({ blockProps }: HeadlineProps) => {
  const { label, size } = blockProps;
  return (
    <h1
      className={twMergeClsx([
        size === "h1" && "text-4xl",
        size === "h2" && "text-3xl",
        size === "h3" && "text-2xl",
      ])}
    >
      <TemplateString value={label} />
    </h1>
  );
};
