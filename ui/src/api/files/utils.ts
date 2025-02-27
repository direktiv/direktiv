import {
  BookOpen,
  File,
  Folder,
  Layers,
  Play,
  Users,
  Workflow,
} from "lucide-react";

import { BaseFileSchemaType } from "./schema";

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

export const fileTypeToIcon = (type: BaseFileSchemaType["type"]) => {
  switch (type) {
    case "directory":
      return Folder;
    case "service":
      return Layers;
    case "workflow":
      return Play;
    case "endpoint":
      return Workflow;
    case "consumer":
      return Users;
    case "gateway":
      return BookOpen;
    default:
      return File;
  }
};

/**
 * This returns the appropriate route with the editor for the specified
 * file type, defaulting to /tree (for directory).
 */
export const fileTypeToExplorerRoute = (type: BaseFileSchemaType["type"]) => {
  switch (type) {
    case "workflow":
      return "/n/$namespace/explorer/workflow/edit/$";
    case "service":
      return "/n/$namespace/explorer/service/$";
    case "endpoint":
      return "/n/$namespace/explorer/endpoint/$";
    case "consumer":
      return "/n/$namespace/explorer/consumer/$";
    case "gateway":
      return "/n/$namespace/explorer/openapiSpecification/$";
    default:
      return "/n/$namespace/explorer/tree/$";
  }
};

export const isPreviewable = (type: BaseFileSchemaType["type"]) =>
  type === "file";
