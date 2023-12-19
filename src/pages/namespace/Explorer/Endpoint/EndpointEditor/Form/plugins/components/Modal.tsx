import {
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { FC, PropsWithChildren } from "react";

import { Button } from "react-day-picker";

export const ModalFooter = () => (
  <DialogFooter>
    <Button type="submit">Save</Button>
  </DialogFooter>
);

type ModalPluginSelectorProps = PropsWithChildren & {
  title: string;
};

export const ModalPluginSelector: FC<ModalPluginSelectorProps> = ({
  title,
  children,
}) => (
  <fieldset className="flex items-center gap-5">
    <label className="text-sm">{title}</label>
    {children}
  </fieldset>
);

type ModalWrapperProps = PropsWithChildren & {
  title: string;
};

export const ModalWrapper: FC<ModalWrapperProps> = ({ title, children }) => (
  <DialogContent className="sm:max-w-xl">
    <DialogHeader>
      <DialogTitle>{title}</DialogTitle>
    </DialogHeader>
    <div className="my-3 flex flex-col gap-5">{children}</div>
  </DialogContent>
);
