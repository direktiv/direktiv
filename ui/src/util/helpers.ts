import { ComponentProps, FC } from "react";
import clsx, { ClassValue } from "clsx";

import { LogEntry } from "~/design/Logs";
import { LogLevelSchemaType } from "~/api/logs/schema";
import moment from "moment";
import { twMerge } from "tailwind-merge";

/**
 * this method combines the usage of two utility libraries:
 * - clsx https://github.com/lukeed/clsx#readme for more info
 *    clsx is a tiny utility for constructing className strings
 *    conditionally.
 * - tailwind-merge https://github.com/dcastil/tailwind-merge
 *    twMerge is a function to efficiently merge Tailwind CSS
 *    classes in JS without style conflicts.
 *
 *    input: twMergeClsx("bg-red", "bg-[#B91C1C]")
 *    output: bg-[#B91C1C]
 */
export const twMergeClsx = (...inputs: ClassValue[]) =>
  twMerge(clsx(...inputs));

type ConditionalWrapperProps = {
  condition: boolean;
  wrapper: (children: JSX.Element) => JSX.Element;
  children: JSX.Element;
};

export const ConditionalWrapper: FC<ConditionalWrapperProps> = ({
  condition,
  wrapper,
  children,
}) => (condition ? wrapper(children) : children);

export const formatLogTime = (time: string) =>
  moment(time).format("YYYY-MM-DD HH:mm:ss.mm");

type LogEntryVariant = ComponentProps<typeof LogEntry>["variant"];

export const logLevelToLogEntryVariant = (
  level: LogLevelSchemaType
): LogEntryVariant => {
  switch (level) {
    case "ERROR":
    case "WARN":
      return "error";
    case "INFO":
      return "info";
    case "DEBUG":
      return undefined;
    default:
      break;
  }
};

export const triggerDownloadFromBase64String = ({
  filename,
  base64String,
  mimeType,
}: {
  filename: string;
  base64String: string;
  mimeType: string;
}) => {
  const aTag = document.createElement("a");
  aTag.href = `data:${mimeType};base64,${base64String}`;
  aTag.download = filename;
  document.body.appendChild(aTag);
  aTag.click();
  document.body.removeChild(aTag);
};

// takes a json input string and format it with 4 spaces indentation
export const prettifyJsonString = (jsonString: string) => {
  try {
    const resultAsJson = JSON.parse(jsonString);
    return JSON.stringify(resultAsJson, null, 4);
  } catch (e) {
    return "{}";
  }
};
