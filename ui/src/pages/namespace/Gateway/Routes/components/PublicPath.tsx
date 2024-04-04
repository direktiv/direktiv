import CopyButton from "~/design/CopyButton";
import { FC } from "react";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";

type PublicPathInputProps = {
  path: string;
};

const PublicPathInput: FC<PublicPathInputProps> = ({ path }) => {
  const absolutePath = `${window.location.origin}${path}`;
  return (
    <InputWithButton
      onClick={(e) => {
        e.stopPropagation(); // prevent the onClick on the row from firing when clicking the workflow link
      }}
    >
      <Input value={absolutePath} readOnly />
      <CopyButton
        value={absolutePath}
        buttonProps={{
          variant: "ghost",
          icon: true,
        }}
      />
    </InputWithButton>
  );
};

export default PublicPathInput;
