import { HeadlineType } from "../../schema/blocks/headline";
import { TemplateString } from "../primitives/TemplateString";
import { twMergeClsx } from "~/util/helpers";

type HeadlineProps = {
  blockProps: HeadlineType;
};

export const Headline = ({ blockProps }: HeadlineProps) => {
  const { label, size } = blockProps;

  const HeadlineTag = size;

  return (
    <HeadlineTag
      className={twMergeClsx([
        size === "h1" && "text-4xl",
        size === "h2" && "text-3xl",
        size === "h3" && "text-2xl",
      ])}
    >
      <TemplateString value={label} />
    </HeadlineTag>
  );
};
