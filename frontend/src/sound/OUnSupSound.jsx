import React from "react";
import { useParams } from "react-router-dom";
import {
  Box,
  Heading,
  Button,
  VStack,
  Text,
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
  Input,
} from "@chakra-ui/react";

// Import Wails runtime
import {
  OpenDirectoryDialog,
  SaveSelectedDirectory,
  GetProjectData,
  ConvertFilesToWAV,
  ProcessAudioChunksAndSpectrograms,
  CalculateOptimalClusters,
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
      processing: false, // Track processing status for spectrograms
      processingMessage: "", // Message for processing status
      duplicates: [], // Track duplicates
      calculatingClusters: false, // Track cluster calculation status
      clustersMessage: "", // Message for cluster calculation status
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

  handleProcessSpectrograms = async () => {
    this.setState({
      processing: true,
      processingMessage: "Processing chunks and generating spectrograms...",
    });
    try {
      const duplicates = await ProcessAudioChunksAndSpectrograms(
        this.props.projectName
      );
      if (duplicates.length > 0) {
        console.log("Duplicates found:", duplicates);
        this.setState({
          processingMessage: `Processing completed with ${duplicates.length} duplicate(s) found.`,
          duplicates,
        });
      } else {
        this.setState({
          processingMessage:
            "Processing completed successfully. No duplicates found.",
        });
      }
    } catch (error) {
      this.setState({
        processingMessage:
          "Error during processing. Please check the console for details.",
      });
      console.error("Error processing spectrograms:", error);
    } finally {
      this.setState({ processing: false });
    }
  };

  handleCalculateClusters = async () => {
    this.setState({
      calculatingClusters: true,
      clustersMessage: "Calculating optimal number of clusters...",
    });
    try {
      const optimalClusters = await CalculateOptimalClusters(
        this.props.projectName
      );
      this.setState({
        clustersMessage: `Optimal number of clusters: ${optimalClusters}`,
      });
    } catch (error) {
      this.setState({
        clustersMessage:
          "Error during cluster calculation. Please check the console for details.",
      });
      console.error("Error calculating optimal clusters:", error);
    } finally {
      this.setState({ calculatingClusters: false });
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
      processing,
      processingMessage,
      duplicates,
      calculatingClusters,
      clustersMessage,
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
                      3. Auto-Detect Chunks & Convert to Spectrograms
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Text>
                      Auto-detect chunks by decibel peaks and convert them to
                      spectrograms:
                    </Text>
                    <Button
                      colorScheme="teal"
                      onClick={this.handleProcessSpectrograms}
                      disabled={processing}
                    >
                      {processing ? "Processing..." : "Start Processing"}
                    </Button>
                    {processingMessage && <Text>{processingMessage}</Text>}
                    {duplicates.length > 0 && (
                      <Text>
                        Duplicates found (see console for details):{" "}
                        {duplicates.join(", ")}
                      </Text>
                    )}
                  </VStack>
                </AccordionPanel>
              </AccordionItem>

              <AccordionItem>
                <h2>
                  <AccordionButton>
                    <Box flex="1" textAlign="left" fontWeight="bold">
                      4. Calculate Optimal Clusters
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                </h2>
                <AccordionPanel pb={4}>
                  <VStack spacing={4} align="start">
                    <Button
                      colorScheme="teal"
                      onClick={this.handleCalculateClusters}
                      disabled={calculatingClusters}
                    >
                      {calculatingClusters
                        ? "Calculating Clusters..."
                        : "Calculate Optimal Clusters"}
                    </Button>
                    {clustersMessage && <Text>{clustersMessage}</Text>}
                  </VStack>
                </AccordionPanel>
              </AccordionItem>

              {/* Step 4: Display Clusters */}
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

              {/* Step 5: Create Model */}
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
