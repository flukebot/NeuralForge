import React from "react";
import { Tabs, TabList, TabPanels, Tab, TabPanel } from "@chakra-ui/react";

class OSupSound extends React.Component {
  state = {};
  render() {
    return (
      <Tabs isFitted variant="enclosed">
        <TabList mb="1em">
          <Tab>One</Tab>
          <Tab>Two</Tab>
        </TabList>
        <TabPanels>
          <TabPanel>
            <p>one!</p>
          </TabPanel>
          <TabPanel>
            <p>two!</p>
          </TabPanel>
        </TabPanels>
      </Tabs>
    );
  }
}

export default OSupSound;
