import { HeadlineType } from "../../schema/blocks/headline";

type HeadlineProps = {
  blockProps: HeadlineType;
};

export const Headline = ({
  blockProps: { label, description },
}: HeadlineProps) => (
  <>
    <h1 className="text-xl">{label}</h1>
    {description && <p>{description}</p>}
  </>
);
