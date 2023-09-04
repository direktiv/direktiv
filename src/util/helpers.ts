import { ComponentProps, FC } from "react";
import clsx, { ClassValue } from "clsx";

import { LogEntry } from "~/design/Logs";
import { LogLevelSchemaType } from "~/api/schema";
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
  moment(time).format("HH:mm:ss.mm");

type LogEntryVariant = ComponentProps<typeof LogEntry>["variant"];

export const logLevelToLogEntryVariant = (
  level: LogLevelSchemaType
): LogEntryVariant => {
  switch (level) {
    case "error":
      return "error";
    case "info":
      return "info";
    case "debug":
      return undefined;
    default:
      break;
  }
};

const mimeTypeToFileExtension = (mimeType: string) => {
  switch (mimeType) {
    case "application/json":
      return ".json";
    case "text/plain":
      return ".txt";
    case "application/x-sh":
      return ".sh";
    case "application/yaml":
    case "text/yaml":
      return ".yaml";
    case "text/html":
      return ".html";
    case "text/xml":
      return ".xml";
    case "text/csv":
      return ".csv";
    default:
      return "";
  }
};

export const triggerDownloadFromBlob = ({
  filename,
  blob,
  mimeType,
}: {
  filename: string;
  blob: Blob;
  mimeType: string;
}) => {
  const url = window.URL.createObjectURL(blob);
  const aTag = document.createElement("a");
  const fileExtension = mimeTypeToFileExtension(mimeType);
  aTag.href = url;
  aTag.download = `${filename}${fileExtension}`;

  document.body.appendChild(aTag);
  aTag.click();
  window.URL.revokeObjectURL(url);
};
