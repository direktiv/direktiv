import { HeadlineType } from "../../schema/blocks/headline";

type HeadlineProps = {
  blockProps: HeadlineType;
};

export const Headline = ({ blockProps: { label } }: HeadlineProps) => (
  <h1 className="text-xl">{label}</h1>
);
