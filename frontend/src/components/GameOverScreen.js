import React, { useState, useEffect } from "react";
import styled from "styled-components";
import { useGame } from "../context/GameContext";

const GameOverContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 2rem;
  text-align: center;
`;

const Title = styled.h1`
  font-size: 3rem;
  margin-bottom: 1rem;
  background: linear-gradient(45deg, #ffd700, #ffed4e);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
`;

const ResultsContainer = styled.div`
  background: rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 2rem;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  max-width: 800px;
  width: 100%;
  margin-bottom: 2rem;
`;

const MetricsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
`;

const MetricCard = styled.div`
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  padding: 1rem;
  text-align: center;
`;

const MetricIcon = styled.div`
  font-size: 2rem;
  margin-bottom: 0.5rem;
`;

const MetricName = styled.div`
  font-size: 0.9rem;
  opacity: 0.8;
  margin-bottom: 0.5rem;
`;

const MetricValue = styled.div`
  font-size: 1.5rem;
  font-weight: bold;
  color: ${(props) => getMetricColor(props.value)};
`;

const StatsSection = styled.div`
  margin-bottom: 2rem;
`;

const StatsTitle = styled.h3`
  color: #ffd700;
  margin-bottom: 1rem;
`;

const StatItem = styled.div`
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
`;

const HistorySection = styled.div`
  text-align: left;
  max-height: 300px;
  overflow-y: auto;
`;

const HistoryTitle = styled.h3`
  color: #ffd700;
  margin-bottom: 1rem;
  text-align: center;
`;

const HistoryItem = styled.div`
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
  border-left: 4px solid ${(props) => getCategoryColor(props.category)};
`;

const TurnNumber = styled.div`
  font-weight: bold;
  color: #ffd700;
  margin-bottom: 0.5rem;
`;

const EventTitle = styled.div`
  font-weight: bold;
  margin-bottom: 0.5rem;
`;

const PlayerChoice = styled.div`
  font-style: italic;
  opacity: 0.8;
  margin-bottom: 0.5rem;
`;

const RestartButton = styled.button`
  background: linear-gradient(45deg, #4caf50, #45a049);
  border: none;
  color: white;
  padding: 1rem 2rem;
  font-size: 1.2rem;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.3);
  }
`;

function getMetricColor(value) {
  if (value >= 50) return "#4CAF50";
  if (value >= 20) return "#8BC34A";
  if (value >= -20) return "#FFC107";
  if (value >= -50) return "#FF9800";
  return "#F44336";
}

function getCategoryColor(category) {
  const colors = {
    economy: "#4CAF50",
    security: "#F44336",
    diplomacy: "#2196F3",
    environment: "#8BC34A",
    domestic: "#FF9800",
    military: "#9C27B0",
    social: "#00BCD4",
    tech: "#607D8B",
  };
  return colors[category] || "#757575";
}

function getMetricIcon(metric) {
  const icons = {
    economy: "💰",
    security: "🛡️",
    diplomacy: "🤝",
    environment: "🌍",
    approval: "👥",
    stability: "⚖️",
  };
  return icons[metric] || "📊";
}

function getMetricName(metric) {
  const names = {
    economy: "Economy",
    security: "Security",
    diplomacy: "Diplomacy",
    environment: "Environment",
    approval: "Approval",
    stability: "Stability",
  };
  return names[metric] || metric;
}

function GameOverScreen({ onRestart }) {
  const { metrics, history, stats, turn, maxTurns } = useGame();
  const [overallScore, setOverallScore] = useState(0);

  useEffect(() => {
    if (metrics) {
      const score =
        Object.values(metrics).reduce((sum, value) => sum + value, 0) / 6;
      setOverallScore(Math.round(score));
    }
  }, [metrics]);

  const getPerformanceText = (score) => {
    if (score >= 50) return { text: "Excellent!", color: "#4CAF50" };
    if (score >= 20) return { text: "Good", color: "#8BC34A" };
    if (score >= -20) return { text: "Satisfactory", color: "#FFC107" };
    if (score >= -50) return { text: "Poor", color: "#FF9800" };
    return { text: "Catastrophic", color: "#F44336" };
  };

  const performance = getPerformanceText(overallScore);

  return (
    <GameOverContainer>
      <Title>Game Over!</Title>

      <ResultsContainer>
        <div style={{ marginBottom: "2rem", textAlign: "center" }}>
          <div style={{ fontSize: "1.5rem", marginBottom: "0.5rem" }}>
            Turns completed: {turn} of {maxTurns}
          </div>
          <div
            style={{
              fontSize: "2rem",
              fontWeight: "bold",
              color: performance.color,
            }}
          >
            {performance.text}
          </div>
          <div style={{ fontSize: "1.2rem", opacity: 0.8 }}>
            Overall Score: {overallScore}
          </div>
        </div>

        <MetricsGrid>
          {metrics &&
            Object.entries(metrics).map(([key, value]) => (
              <MetricCard key={key}>
                <MetricIcon>{getMetricIcon(key)}</MetricIcon>
                <MetricName>{getMetricName(key)}</MetricName>
                <MetricValue value={value}>
                  {value > 0 ? "+" : ""}
                  {Math.round(value)}
                </MetricValue>
              </MetricCard>
            ))}
        </MetricsGrid>

        {stats && (
          <StatsSection>
            <StatsTitle>AI Usage Statistics</StatsTitle>
            <StatItem>
              <span>Advisors (Theta):</span>
              <span>{stats.advisorTheta}</span>
            </StatItem>
            <StatItem>
              <span>Advisors (Gemini):</span>
              <span>{stats.advisorGemini}</span>
            </StatItem>
            <StatItem>
              <span>Director (Theta):</span>
              <span>{stats.directorTheta}</span>
            </StatItem>
            <StatItem>
              <span>Director (Gemini):</span>
              <span>{stats.directorGemini}</span>
            </StatItem>
            <StatItem>
              <span>Rewriting (Gemini):</span>
              <span>{stats.rewriteGemini}</span>
            </StatItem>
          </StatsSection>
        )}

        {history && history.length > 0 && (
          <HistorySection>
            <HistoryTitle>Decision History</HistoryTitle>
            {history.map((turn, index) => (
              <HistoryItem key={index} category={turn.event?.category}>
                <TurnNumber>Turn {turn.turn}</TurnNumber>
                <EventTitle>{turn.event?.title}</EventTitle>
                <PlayerChoice>Choice: {turn.choice?.option}</PlayerChoice>
                {turn.evaluation && (
                  <div style={{ fontSize: "0.9rem", opacity: 0.8 }}>
                    {turn.evaluation}
                  </div>
                )}
              </HistoryItem>
            ))}
          </HistorySection>
        )}
      </ResultsContainer>

      <RestartButton onClick={onRestart}>Start New Game</RestartButton>
    </GameOverContainer>
  );
}

export default GameOverScreen;
