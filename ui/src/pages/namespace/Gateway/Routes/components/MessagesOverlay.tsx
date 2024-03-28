import Alert, { AlertProps } from "~/design/Alert";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

import { FC } from "react";

export type MessagesOverlayProps = {
  messages: string[];
  children: (messages: number) => JSX.Element;
  variant?: AlertProps["variant"];
};

const MessagesOverlay: FC<MessagesOverlayProps> = ({
  messages,
  children,
  variant,
}) => {
  const messageCount = messages.length;

  return messageCount > 0 ? (
    <HoverCard>
      <HoverCardTrigger className="inline-flex">
        {children(messageCount)}
      </HoverCardTrigger>
      <HoverCardContent className="flex flex-col gap-2 p-1">
        {messages.map((message, i) => (
          <Alert
            key={i}
            variant={variant}
            className="w-96 whitespace-pre-wrap break-all"
          >
            {message}
          </Alert>
        ))}
      </HoverCardContent>
    </HoverCard>
  ) : null;
};

export default MessagesOverlay;
