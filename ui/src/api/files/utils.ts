import {
  Braces,
  File,
  Folder,
  Layers,
  LucideIcon,
  Play,
  Users,
  Workflow,
} from "lucide-react";

import { BaseFileSchemaType } from "./schema";
import { ExplorerSubpages } from "~/util/router/pages";

export const getFilenameFromPath = (path: string): string => {
  const fileName = path.split("/").pop();
  if (fileName === undefined)
    throw Error(`Filename could not be extracted from ${path}`);
  return fileName;
};

export const getParentFromPath = (path: string): string => {
  switch (path) {
    case "":
      throw Error("Cannot infer parent from empty string");
    case "/":
      throw Error("Cannot infer parent from '/'");
    default:
      return path.split("/").slice(0, -1).join("/") || "/";
  }
};

export const sortFoldersFirst = (
  a: BaseFileSchemaType,
  b: BaseFileSchemaType
): number => {
  if (a.type === "directory" && b.type !== "directory") {
    return -1;
  }

  if (b.type === "directory" && a.type !== "directory") {
    return 1;
  }

  return a.path.localeCompare(b.path);
};

export const forceLeadingSlash = (path?: string) => {
  if (!path) {
    return "/";
  }
  return path.startsWith("/") ? path : `/${path}`;
};

export const removeLeadingSlash = (path?: string) => {
  if (!path) {
    return "";
  }
  return path.startsWith("/") ? path.slice(1) : path;
};

export const removeTrailingSlash = (path?: string) => {
  if (!path) {
    return "";
  }
  return path.endsWith("/") ? path.slice(0, -1) : path;
};

export const sortByName = (a: { name: string }, b: { name: string }): number =>
  a.name.localeCompare(b.name);

export const workflowIcon = (mimeType: string | undefined): LucideIcon => {
  switch (mimeType) {
    case "application/x-typescript":
      return Braces;
    default:
      return Play;
  }
};

export const fileTypeToIcon = (
  type: BaseFileSchemaType["type"],
  mimeType: string | undefined
) => {
  switch (type) {
    case "directory":
      return Folder;
    case "service":
      return Layers;
    case "workflow":
      return workflowIcon(mimeType);
    case "endpoint":
      return Workflow;
    case "consumer":
      return Users;
    default:
      return File;
  }
};

export const fileTypeToExplorerSubpage = (
  type: BaseFileSchemaType["type"]
): ExplorerSubpages | undefined => {
  switch (type) {
    case "workflow":
      return "workflow";
    case "service":
      return "service";
    case "endpoint":
      return "endpoint";
    case "consumer":
      return "consumer";
    default:
      return undefined;
  }
};

export const isPreviewable = (type: BaseFileSchemaType["type"]) =>
  type === "file";
