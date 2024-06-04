import { ComponentType, PropsWithChildren, ReactNode } from "react";

type WrapperComponent = ComponentType<{
  onClick: () => void;
  children: ReactNode;
}>;

export type LogoutButtonProps = PropsWithChildren & {
  wrapper: WrapperComponent;
};
