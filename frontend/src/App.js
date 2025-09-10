import React, { useState } from "react";
import { GameProvider } from "./context/GameContext";
import { useGame } from "./context/GameContext";
import StartScreen from "./components/StartScreen";
import GameBoard from "./components/GameBoard";
import GameOverScreen from "./components/GameOverScreen";
import GameSelector from "./components/GameSelector";
import GameDescription from "./components/GameDescription";
import styled from "styled-components";
import "./App.css";

const AppContainer = styled.div`
  min-height: 100vh;
  background: #0f1419;
  display: flex;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  width: 100vw;
  overflow-x: hidden;
`;

const MainContent = styled.div`
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  overflow: hidden;
  margin-left: 0;

  @media (min-width: 769px) {
    margin-left: 0;
  }
`;

function AppContent() {
  const { gameState, startGame } = useGame();
  const [selectedGame, setSelectedGame] = useState("presidential-simulator");

  const handleGameSelect = (gameId) => {
    setSelectedGame(gameId);
  };

  const handleStartGame = () => {
    if (selectedGame === "presidential-simulator") {
      startGame();
    }
  };

  return (
    <AppContainer>
      <GameSelector
        selectedGame={selectedGame}
        onGameSelect={handleGameSelect}
      />
      <MainContent>
        {gameState === "start" && (
          <GameDescription
            gameId={selectedGame}
            onStartGame={handleStartGame}
          />
        )}
        {gameState === "playing" && <GameBoard />}
        {gameState === "gameOver" && <GameOverScreen />}
      </MainContent>
    </AppContainer>
  );
}

function App() {
  return (
    <GameProvider>
      <AppContent />
    </GameProvider>
  );
}

export default App;
