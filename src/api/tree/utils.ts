import { File, Folder, Layers, Network, Play, Users } from "lucide-react";

import { ExplorerSubpages } from "~/util/router/pages";
import { NodeSchemaType } from "./schema/node";

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

export const sortFoldersFirst = (
  a: NodeSchemaType,
  b: NodeSchemaType
): number => {
  if (a.type === "directory" && b.type !== "directory") {
    return -1;
  }

  if (b.type === "directory" && a.type !== "directory") {
    return 1;
  }

  return a.name.localeCompare(b.name);
};

export const sortByName = (a: { name: string }, b: { name: string }): number =>
  a.name.localeCompare(b.name);

export const sortByRef = (a: { ref: string }, b: { ref: string }): number =>
  a.ref.localeCompare(b.ref);

export const fileTypeToIcon = (type: NodeSchemaType["type"]) => {
  switch (type) {
    case "directory":
      return Folder;
    case "service":
      return Layers;
    case "workflow":
      return Play;
    case "endpoint":
      return Network;
    case "consumer":
      return Users;
    default:
      return File;
  }
};

export const fileTypeToExplorerSubpage = (
  type: NodeSchemaType["type"]
): ExplorerSubpages | undefined => {
  switch (type) {
    case "workflow":
      return "workflow";
    case "service":
      return "service";
    case "endpoint":
      return "workflow";
    case "consumer":
      return "workflow";
    default:
      return undefined;
  }
};

export const isPreviewable = (type: NodeSchemaType["type"]) => type === "file";
