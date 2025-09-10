import React from "react";
import styled from "styled-components";

const EventContainer = styled.div`
  background: #1a1a1a;
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1rem 0;
  border: 1px solid #333;
  border-left: 4px solid #ff4444;
`;

const NewsHeader = styled.div`
  color: #4a9eff;
  font-size: 0.9rem;
  margin-bottom: 1rem;
  font-weight: 500;
`;

const EventTitle = styled.h3`
  color: #ff4444;
  font-size: 1.1rem;
  margin: 0.5rem 0;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.5rem;
`;

const EventDescription = styled.div`
  color: #ffffff;
  line-height: 1.6;
  margin: 1rem 0;
  font-size: 0.95rem;
  white-space: pre-wrap;
`;

const EventSection = styled.div`
  margin: 0.8rem 0;
  color: #cccccc;
  font-size: 0.9rem;
  white-space: pre-wrap;

  strong {
    color: #4a9eff;
  }
`;

const HashTags = styled.div`
  margin-top: 1rem;
  color: #4a9eff;
  font-size: 0.85rem;
  font-style: italic;
`;

function EventDisplay({ event }) {
  if (!event) return null;

  return (
    <EventContainer>
      <NewsHeader>
        📧 [BREAKING NEWS] New situation requires your immediate attention...
      </NewsHeader>

      <EventTitle>🚨 🚨 BREAKING: {event.title}</EventTitle>

      <EventDescription>
        📝 🧭 <strong>What's Happening:</strong> {event.description} (severity{" "}
        {event.severity}/10).
      </EventDescription>

      {event.onTheGround && (
        <EventSection>
          🌐 <strong>On The Ground:</strong> {event.onTheGround}
        </EventSection>
      )}

      {event.namedEntity && (
        <EventSection>
          🏢 <strong>Named Entity:</strong> {event.namedEntity}
        </EventSection>
      )}

      {event.officialVsRumors && (
        <EventSection>
          🗣️ <strong>Official vs Rumors:</strong> {event.officialVsRumors}
        </EventSection>
      )}

      {event.politics && (
        <EventSection>
          🏛️ <strong>Politics:</strong> {event.politics}
        </EventSection>
      )}

      {event.hashtags && <HashTags>{event.hashtags}</HashTags>}
    </EventContainer>
  );
}

export default EventDisplay;
