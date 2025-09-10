import React from "react";
import styled from "styled-components";

const Panel = styled.div`
  background: #212121;
  border-radius: 12px;
  padding: 1rem;
  border: 1px solid #2f2f2f;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
`;

const Title = styled.h3`
  margin: 0 0 1rem 0;
  text-align: center;
  color: #ffffff;
  font-size: 1rem;
  font-weight: 500;
`;

const MetricItem = styled.div`
  margin-bottom: 0.8rem;
  padding: 0.5rem;
  background: #2a2a2a;
  border-radius: 8px;
  border: 1px solid #3a3a3a;
`;

const MetricLabel = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  font-size: 0.85rem;
  color: #ffffff;
`;

const MetricBar = styled.div`
  height: 6px;
  background: #3a3a3a;
  border-radius: 3px;
  overflow: hidden;
`;

const MetricFill = styled.div`
  height: 100%;
  background: ${(props) => getMetricColor(props.value)};
  width: ${(props) => Math.max(0, Math.min(100, (props.value + 100) / 2))}%;
  transition: all 0.3s ease;
`;

const MetricValue = styled.span`
  font-weight: 500;
  color: ${(props) => getMetricColor(props.value)};
  font-size: 0.8rem;
`;

function getMetricColor(value) {
  if (value >= 50) return "#4CAF50";
  if (value >= 20) return "#8BC34A";
  if (value >= -20) return "#FFC107";
  if (value >= -50) return "#FF9800";
  return "#F44336";
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

function MetricsPanel({ metrics }) {
  if (!metrics) return null;

  const metricKeys = [
    "economy",
    "security",
    "diplomacy",
    "environment",
    "approval",
    "stability",
  ];

  return (
    <Panel>
      <Title>Country Status</Title>
      {metricKeys.map((key) => {
        const value = metrics[key] || 0;
        return (
          <MetricItem key={key}>
            <MetricLabel>
              <span>
                {getMetricIcon(key)} {getMetricName(key)}
              </span>
              <MetricValue value={value}>
                {value > 0 ? "+" : ""}
                {Math.round(value)}
              </MetricValue>
            </MetricLabel>
            <MetricBar>
              <MetricFill value={value} />
            </MetricBar>
          </MetricItem>
        );
      })}
    </Panel>
  );
}

export default MetricsPanel;
