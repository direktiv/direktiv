import { HeadlineType } from "../../schema/blocks/headline";
import { TemplateString } from "../primitives/TemplateString";
import { twMergeClsx } from "~/util/helpers";

type HeadlineProps = {
  blockProps: HeadlineType;
};

export const Headline = ({ blockProps }: HeadlineProps) => {
  const { label, level } = blockProps;

  const HeadlineTag = level;

  return (
    <HeadlineTag
      className={twMergeClsx([
        level === "h1" && "text-4xl",
        level === "h2" && "text-3xl",
        level === "h3" && "text-2xl",
      ])}
    >
      <TemplateString value={label} />
    </HeadlineTag>
  );
};
