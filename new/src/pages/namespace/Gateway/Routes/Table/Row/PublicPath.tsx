import CopyButton from "~/design/CopyButton";
import { FC } from "react";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";

type PublicPathInputProps = {
  path: string;
};

const PublicPathInput: FC<PublicPathInputProps> = ({ path }) => (
  <InputWithButton>
    <Input value={path} readOnly />
    <CopyButton
      value={path}
      buttonProps={{
        variant: "ghost",
        icon: true,
      }}
    />
  </InputWithButton>
);

export default PublicPathInput;
