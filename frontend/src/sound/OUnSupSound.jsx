import React from "react";
import { useParams } from "react-router-dom";
import {
  Box,
  Heading,
  Button,
  VStack,
  Text,
  Input,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Accordion,
  AccordionItem,
  AccordionButton,
  AccordionPanel,
  AccordionIcon,
  Divider,
  Stack,
} from "@chakra-ui/react";

// Import Wails runtime
import {
  OpenDirectoryDialog,
  SaveSelectedDirectory,
  GetProjectData,
  ConvertFilesToWAV,
} from "../../wailsjs/go/main/App";

class OUnSupSound extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      loading: true,
      selectedFolderPath: "",
      fileList: {},
      errorMessage: "",
      converting: false, // Track conversion status
      conversionMessage: "", // Message for conversion status
    };
  }

  async componentDidMount() {
    await this.fetchProjectData();
  }

  fetchProjectData = async () => {
    try {
      const projectData = await GetProjectData(this.props.projectName);
      this.setState({
        selectedFolderPath: projectData.selected_directory,
        fileList: projectData.file_list,
        loading: false,
      });
    } catch (error) {
      this.setState({
        loading: false,
        errorMessage:
          "No existing configuration found. Please select a folder.",
      });
    }
  };

  handleSelectFolder = async () => {
    try {
      const selectedPath = await OpenDirectoryDialog();
      console.log("Selected Path:", selectedPath);

      if (selectedPath) {
        await SaveSelectedDirectory(selectedPath, this.props.projectName);
        this.setState({ selectedFolderPath: selectedPath });
        await this.fetchProjectData(); // Re-fetch to update file list
      }
    } catch (error) {
      console.error("Error selecting folder:", error);
      this.setState({
        errorMessage: "Error selecting folder. Please try again.",
      });
    }
  };

  handleStartConversion = async () => {
    this.setState({
      converting: true,
      conversionMessage: "Conversion in progress...",
    });
    try {
      await ConvertFilesToWAV(this.props.projectName);
      this.setState({
        conversionMessage: "Conversion completed successfully.",
      });
    } catch (error) {
      this.setState({
        conversionMessage:
          "Error during conversion. Please check the console for details.",
      });
      console.error("Error converting files to WAV:", error);
    } finally {
      this.setState({ converting: false });
    }
  };

  render() {
    const { projectName } = this.props;
    const {
      loading,
      selectedFolderPath,
      errorMessage,
      converting,
      conversionMessage,
    } = this.state;

    return (
      <Box p={5}>
        <Heading as="h1" size="xl" mb={5} color="teal.600">
          Unsupervised Learning - {projectName}
        </Heading>

        {loading ? (
          <Spinner size="xl" />
        ) : (
          <VStack spacing={4} align="start">
            {errorMessage && (
              <Alert status="error">
                <AlertIcon />
                <AlertTitle>Error:</AlertTitle>
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>
            )}

            <Accordion allowToggle defaultIndex={[0]}>
              {/* Step 1: Select Folder */}
              <AccordionItem>
                <h2>
                  <AccordionButton>
                    <Box flex="1" textAlign="left" fontWeight="bold">
                      1. Select Folder
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Text>
                      Select the folder containing audio files and subfiles:
                    </Text>
                    <Button
                      colorScheme="teal"
                      onClick={this.handleSelectFolder}
                    >
                      {selectedFolderPath ? "Change Folder" : "Choose Folder"}
                    </Button>
                    {selectedFolderPath && (
                      <Text>Selected folder: {selectedFolderPath}</Text>
                    )}
                  </VStack>
                </AccordionPanel>
              </AccordionItem>

              {/* Step 2: Convert Files */}
              <AccordionItem>
                <h2>
                  <AccordionButton>
                    <Box flex="1" textAlign="left" fontWeight="bold">
                      2. Convert Files
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Text>Convert all files to WAV format using FFmpeg:</Text>
                    <Button
                      colorScheme="teal"
                      onClick={this.handleStartConversion}
                      disabled={converting}
                    >
                      {converting ? "Converting..." : "Start Conversion"}
                    </Button>
                    {conversionMessage && <Text>{conversionMessage}</Text>}
                  </VStack>
                </AccordionPanel>
              </AccordionItem>

              {/* Step 3: Auto-Detect Chunks */}
              <AccordionItem>
                <h2>
                  <AccordionButton>
                    <Box flex="1" textAlign="left" fontWeight="bold">
                      3. Auto-Detect Chunks
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Text>
                      Auto-detect and chunk audio based on decibel ratings.
                      Specify chunking strategy:
                    </Text>
                    <Stack direction="row" spacing={4}>
                      <Button colorScheme="teal">Chunk by Peaks</Button>
                      <Button colorScheme="teal">Chunk by Time</Button>
                    </Stack>
                    <Text>
                      Detection and chunking progress will be shown here.
                    </Text>
                  </VStack>
                </AccordionPanel>
              </AccordionItem>

              {/* Step 4: Convert to Spectrograms */}
              <AccordionItem>
                <h2>
                  <AccordionButton>
                    <Box flex="1" textAlign="left" fontWeight="bold">
                      4. Convert to Spectrograms
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Text>Convert audio chunks to spectrograms:</Text>
                    <Button colorScheme="teal">Convert to Spectrograms</Button>
                    <Text>
                      Spectrogram generation progress or completion will be
                      shown here.
                    </Text>
                  </VStack>
                </AccordionPanel>
              </AccordionItem>

              {/* Step 5: Display Clusters */}
              <AccordionItem>
                <h2>
                  <AccordionButton>
                    <Box flex="1" textAlign="left" fontWeight="bold">
                      5. Display Clusters
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Text>Diagrammatically display all the clusters:</Text>
                    <Button colorScheme="teal">Show Clusters</Button>
                    <Box
                      w="100%"
                      h="300px"
                      bg="gray.100"
                      borderRadius="md"
                      p={4}
                    >
                      {/* Placeholder for cluster diagram */}
                      Cluster diagram will be displayed here.
                    </Box>
                  </VStack>
                </AccordionPanel>
              </AccordionItem>

              {/* Step 6: Create Model */}
              <AccordionItem>
                <h2>
                  <AccordionButton>
                    <Box flex="1" textAlign="left" fontWeight="bold">
                      6. Create Model
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Text>Create TensorFlow.js model:</Text>
                    <Text>
                      Specify the number of layers and model type (e.g., CNN):
                    </Text>
                    <Input placeholder="Number of layers" type="number" />
                    <Button colorScheme="teal">Create Model</Button>
                    <Text>Model creation status will be displayed here.</Text>
                  </VStack>
                </AccordionPanel>
              </AccordionItem>
            </Accordion>
          </VStack>
        )}
      </Box>
    );
  }
}

const OUnSupSoundWithParams = (props) => {
  const params = useParams();
  return <OUnSupSound {...props} projectName={params.projectName} />;
};

export default OUnSupSoundWithParams;
