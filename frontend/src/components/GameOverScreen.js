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
  background: radial-gradient(60% 80% at 50% 20%, rgba(13, 71, 161, 0.35), rgba(10, 25, 47, 0.9)), #0a192f;
  color: #e3f2fd;
`;

const Title = styled.h1`
  font-size: 3rem;
  margin-bottom: 1rem;
  background: linear-gradient(45deg, #4fc3f7, #0288d1);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 2px 2px 12px rgba(2, 136, 209, 0.2);
`;

const ResultsContainer = styled.div`
  background: rgba(13, 71, 161, 0.15);
  border-radius: 20px;
  padding: 2rem;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(79, 195, 247, 0.25);
  max-width: 900px;
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
  background: rgba(79, 195, 247, 0.08);
  border-radius: 10px;
  padding: 1rem;
  text-align: center;
  border: 1px solid rgba(79, 195, 247, 0.18);
`;

const MetricIcon = styled.div`
  font-size: 2rem;
  margin-bottom: 0.5rem;
`;

const MetricName = styled.div`
  font-size: 0.9rem;
  opacity: 0.9;
  margin-bottom: 0.5rem;
  color: #bbdefb;
`;

const MetricValue = styled.div`
  font-size: 1.5rem;
  font-weight: bold;
  color: ${(props) => getMetricColor(props.value)};
`;

const HistorySection = styled.div`
  text-align: left;
  max-height: 340px;
  overflow-y: auto;
`;

const HistoryTitle = styled.h3`
  color: #4fc3f7;
  margin-bottom: 1rem;
  text-align: center;
`;

const HistoryItem = styled.div`
  background: rgba(79, 195, 247, 0.06);
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
  border-left: 4px solid ${(props) => getCategoryColor(props.category)};
`;

const TurnNumber = styled.div`
  font-weight: bold;
  color: #81d4fa;
  margin-bottom: 0.5rem;
`;

const EventTitle = styled.div`
  font-weight: bold;
  margin-bottom: 0.5rem;
`;

const PlayerChoice = styled.div`
  font-style: italic;
  opacity: 0.9;
  margin-bottom: 0.5rem;
  color: #b3e5fc;
`;

const RestartButton = styled.button`
  background: linear-gradient(45deg, #42a5f5, #1e88e5);
  border: none;
  color: white;
  padding: 1rem 2rem;
  font-size: 1.2rem;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 6px 20px rgba(2, 136, 209, 0.35);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 28px rgba(2, 136, 209, 0.5);
  }
`;

function getMetricColor(value) {
  if (value >= 50) return "#81C784"; // soft green
  if (value >= 20) return "#AED581";
  if (value >= -20) return "#FFEE58"; // yellow
  if (value >= -50) return "#FFB74D"; // orange
  return "#E57373"; // soft red
}

function getCategoryColor(category) {
  const colors = {
    economy: "#26A69A",
    security: "#42A5F5",
    diplomacy: "#5C6BC0",
    environment: "#66BB6A",
    domestic: "#29B6F6",
    military: "#7E57C2",
    social: "#26C6DA",
    tech: "#90A4AE",
  };
  return colors[category] || "#4FC3F7";
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
  const { metrics, history, turn, maxTurns, startGame } = useGame();
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
  const shownTurn = Math.min(turn, maxTurns);

  const handleRestart = async () => {
    try {
      await startGame();
    } catch (e) {
      // no-op
    }
  };

  return (
    <GameOverContainer>
      <Title>Game Over!</Title>

      <ResultsContainer>
        <div style={{ marginBottom: "2rem", textAlign: "center" }}>
          <div style={{ fontSize: "1.5rem", marginBottom: "0.5rem" }}>
            Turns completed: {shownTurn} of {maxTurns}
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
          <div style={{ fontSize: "1.2rem", opacity: 0.85 }}>
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

        {history && history.length > 0 && (
          <HistorySection>
            <HistoryTitle>Decision History</HistoryTitle>
            {history.map((turn, index) => (
              <HistoryItem key={index} category={turn.event?.category}>
                <TurnNumber>Turn {turn.turn}</TurnNumber>
                <EventTitle>{turn.event?.title}</EventTitle>
                <PlayerChoice>
                  Choice: {turn.choice?.reasoning || turn.choice?.option}
                </PlayerChoice>
                {turn.evaluation && (
                  <div style={{ fontSize: "0.9rem", opacity: 0.9 }}>
                    {turn.evaluation}
                  </div>
                )}
              </HistoryItem>
            ))}
          </HistorySection>
        )}
      </ResultsContainer>

      <RestartButton onClick={onRestart || handleRestart}>Start New Game</RestartButton>
    </GameOverContainer>
  );
}

export default GameOverScreen;
