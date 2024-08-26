import React from "react";
import { useParams } from "react-router-dom";

import {
  Box,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Heading,
  Button,
  VStack,
  Input,
  Text,
  Stack,
} from "@chakra-ui/react";

class OUnSupSound extends React.Component {
  render() {
    const { projectName } = this.props;
    console.log("Received Project Name:", projectName);

    return (
      <Box p={5}>
        <Heading as="h1" size="xl" mb={5} color="teal.600">
          Unsupervised Learning - {projectName}
        </Heading>
        <Tabs variant="enclosed" colorScheme="teal">
          <TabList>
            <Tab>Select Folder</Tab>
            <Tab>Convert Files</Tab>
            <Tab>Auto-Detect Chunks</Tab>
            <Tab>Convert to Spectrograms</Tab>
            <Tab>Display Clusters</Tab>
            <Tab>Create Model</Tab>
          </TabList>

          <TabPanels>
            {/* Step 1: Select Folder */}
            <TabPanel>
              <VStack spacing={4} align="start">
                <Text>
                  Select the folder containing audio files and subfiles:
                </Text>
                <Button colorScheme="teal">Choose Folder</Button>
                <Text>Selected folder path will be displayed here.</Text>
              </VStack>
            </TabPanel>

            {/* Step 2: Convert Files */}
            <TabPanel>
              <VStack spacing={4} align="start">
                <Text>Convert all files to WAV format using FFmpeg:</Text>
                <Button colorScheme="teal">Start Conversion</Button>
                <Text>
                  Conversion progress or completion message will be displayed
                  here.
                </Text>
              </VStack>
            </TabPanel>

            {/* Step 3: Auto-Detect Chunks */}
            <TabPanel>
              <VStack spacing={4} align="start">
                <Text>
                  Auto-detect and chunk audio based on decibel ratings. Specify
                  chunking strategy:
                </Text>
                <Stack direction="row" spacing={4}>
                  <Button colorScheme="teal">Chunk by Peaks</Button>
                  <Button colorScheme="teal">Chunk by Time</Button>
                </Stack>
                <Text>Detection and chunking progress will be shown here.</Text>
              </VStack>
            </TabPanel>

            {/* Step 4: Convert to Spectrograms */}
            <TabPanel>
              <VStack spacing={4} align="start">
                <Text>Convert audio chunks to spectrograms:</Text>
                <Button colorScheme="teal">Convert to Spectrograms</Button>
                <Text>
                  Spectrogram generation progress or completion will be shown
                  here.
                </Text>
              </VStack>
            </TabPanel>

            {/* Step 5: Display Clusters */}
            <TabPanel>
              <VStack spacing={4} align="start">
                <Text>Diagrammatically display all the clusters:</Text>
                <Button colorScheme="teal">Show Clusters</Button>
                <Box w="100%" h="300px" bg="gray.100" borderRadius="md" p={4}>
                  {/* Placeholder for cluster diagram */}
                  Cluster diagram will be displayed here.
                </Box>
              </VStack>
            </TabPanel>

            {/* Step 6: Create Model */}
            <TabPanel>
              <VStack spacing={4} align="start">
                <Text>Create TensorFlow.js model:</Text>
                <Text>
                  Specify the number of layers and model type (e.g., CNN):
                </Text>
                <Input placeholder="Number of layers" type="number" />
                <Button colorScheme="teal">Create Model</Button>
                <Text>Model creation status will be displayed here.</Text>
              </VStack>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </Box>
    );
  }
}

const OUnSupSoundWithParams = (props) => {
  const params = useParams();
  return <OUnSupSound {...props} projectName={params.projectName} />;
};

export default OUnSupSoundWithParams;
