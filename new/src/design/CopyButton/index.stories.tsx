import CopyButton from ".";

export default {
  title: "Components/CopyButton",
  parameters: { layout: "fullscreen" },
};

export const Default = () => (
  <div className="p-5">
    <CopyButton value="copy me" />
  </div>
);

export const StyledButton = () => (
  <div className="p-5">
    <CopyButton
      value="copy me"
      buttonProps={{
        variant: "primary",
        size: "sm",
      }}
    />
  </div>
);

export const WithRenderProps = () => (
  <div className="p-5">
    <CopyButton
      value="copy me"
      buttonProps={{
        variant: "outline",
        className: "w-40",
      }}
    >
      {(copied) => (copied ? "Copied" : "Copy")}
    </CopyButton>
  </div>
);
