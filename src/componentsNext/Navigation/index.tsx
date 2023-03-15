import {
  Box,
  Bug,
  Calendar,
  FolderTree,
  Layers,
  Network,
  Settings,
  Users,
} from "lucide-react";

import { FC } from "react";
import { NavigationLink } from "../NavigationLink";

export const navigation = [
  { name: "Explorer", href: "#", icon: FolderTree, current: true },
  { name: "Monitoring", href: "#", icon: Bug, current: false },
  { name: "Instances", href: "#", icon: Box, current: false },
  { name: "Events", href: "#", icon: Calendar, current: false },
  { name: "Gateway", href: "#", icon: Network, current: false },
  { name: "Permissions", href: "#", icon: Users, current: false },
  { name: "Services", href: "#", icon: Layers, current: false },
  { name: "Settings", href: "#", icon: Settings, current: false },
];

const Navigation: FC = () => (
  <>
    {navigation.map((item) => (
      <NavigationLink key={item.name} href={item.href} active={item.current}>
        <item.icon aria-hidden="true" />
        {item.name}
      </NavigationLink>
    ))}
  </>
);

export default Navigation;
