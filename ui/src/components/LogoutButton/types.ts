import { ComponentType, PropsWithChildren, ReactNode } from "react";

type ButtonComponent = ComponentType<{
  onClick: () => void;
  children: ReactNode;
}>;

export type LogoutButtonProps = PropsWithChildren & {
  button: ButtonComponent;
};
