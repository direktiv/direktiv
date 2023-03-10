import {
  Main,
  MainContent,
  MainTop,
  MainTopLeft,
  MainTopRight,
  Root,
  Sidebar,
  SidebarMain,
  SidebarTop,
} from "./index";

export default {
  title: "Components (next)/AppShell",
  parameters: { layout: "fullscreen" },
};

export const Default = () => (
  <Root>
    <Sidebar version="version">
      <SidebarTop>
        <div className="lg:hidden">Burger Menu Button</div>
      </SidebarTop>
      <SidebarMain>Sidebar</SidebarMain>
    </Sidebar>
    <Main>
      <MainTop>
        <MainTopLeft>Top Left</MainTopLeft>
        <MainTopRight>Top Right</MainTopRight>
      </MainTop>
      <MainContent>
        <div className="p-10">Main Content</div>
      </MainContent>
    </Main>
  </Root>
);
