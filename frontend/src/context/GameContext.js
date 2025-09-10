import React, { createContext, useContext, useReducer } from "react";
import axios from "axios";

const GameContext = createContext();

const initialState = {
  gameState: "start",
  turn: 0,
  maxTurns: 5,
  metrics: {
    economy: 0,
    security: 0,
    diplomacy: 0,
    environment: 0,
    approval: 0,
    stability: 0,
  },
  currentTurn: null,
  history: [],
  advisors: [],
  isComplete: false,
  loading: false,
  error: null,
  stats: {
    advisorTheta: 0,
    advisorGemini: 0,
    directorTheta: 0,
    directorGemini: 0,
    rewriteGemini: 0,
  },
};

function gameReducer(state, action) {
  switch (action.type) {
    case "SET_LOADING":
      return { ...state, loading: action.payload };
    case "SET_ERROR":
      return { ...state, error: action.payload, loading: false };
    case "START_GAME":
      return { ...initialState, gameState: "playing", loading: false };
    case "UPDATE_GAME_STATE":
      return {
        ...state,
        ...action.payload,
        loading: false,
        error: null,
      };
    case "NEW_ROUND":
      return {
        ...state,
        gameState: "playing",
        turn: action.payload.turn,
        maxTurns: action.payload.maxTurns,
        currentTurn: action.payload.turnResult,
        metrics: action.payload.metrics || state.metrics,
        stats: action.payload.stats,
        loading: false,
      };
    case "EVALUATE_CHOICE":
      return {
        ...state,
        gameState: action.payload.isComplete ? "gameOver" : "playing",
        metrics: action.payload.metrics,
        isComplete: action.payload.isComplete,
        turn: action.payload.turn,
        maxTurns: action.payload.maxTurns,
        stats: action.payload.stats,
        history: [
          ...state.history,
          {
            ...state.currentTurn,
            choice: action.payload.choice,
            evaluation: action.payload.evaluation,
            impact: action.payload.impact,
          },
        ],
        currentTurn: null,
        loading: false,
      };
    default:
      return state;
  }
}

export function GameProvider({ children }) {
  const [state, dispatch] = useReducer(gameReducer, initialState);

  const api = axios.create({
    baseURL: "http://localhost:8080/api",
    timeout: 30000,
  });

  api.interceptors.request.use(
    (config) => {
      console.log('🚀 API Request:', {
        method: config.method?.toUpperCase(),
        url: config.url,
        baseURL: config.baseURL,
        fullURL: `${config.baseURL}${config.url}`,
        data: config.data,
        params: config.params,
        headers: config.headers,
        timestamp: new Date().toISOString()
      });
      return config;
    },
    (error) => {
      console.error('❌ API Request Error:', error);
      return Promise.reject(error);
    }
  );

  api.interceptors.response.use(
    (response) => {
      console.log('✅ API Response:', {
        status: response.status,
        statusText: response.statusText,
        url: response.config.url,
        method: response.config.method?.toUpperCase(),
        data: response.data,
        headers: response.headers,
        timestamp: new Date().toISOString()
      });
      return response;
    },
    (error) => {
      console.error('❌ API Response Error:', {
        message: error.message,
        status: error.response?.status,
        statusText: error.response?.statusText,
        url: error.config?.url,
        method: error.config?.method?.toUpperCase(),
        data: error.response?.data,
        timestamp: new Date().toISOString()
      });
      return Promise.reject(error);
    }
  );

  const startGame = async () => {
    dispatch({ type: "SET_LOADING", payload: true });
    try {
      await api.post("/start");
      dispatch({ type: "START_GAME" });
    } catch (error) {
      dispatch({ type: "SET_ERROR", payload: error.message });
    }
  };

  const newRound = async () => {
    dispatch({ type: "SET_LOADING", payload: true });
    try {
      const response = await api.post("/new-round");
      dispatch({ type: "NEW_ROUND", payload: response.data });
      return response.data;
    } catch (error) {
      dispatch({ type: "SET_ERROR", payload: error.message });
      throw error;
    }
  };

  const evaluateChoice = async (eventId, optionIndex, option, reasoning) => {
    dispatch({ type: "SET_LOADING", payload: true });
    try {
      const response = await api.post("/evaluate-choice", {
        eventId,
        optionIndex,
        option,
        reasoning,
      });
      dispatch({
        type: "EVALUATE_CHOICE",
        payload: {
          ...response.data,
          choice: { eventId, optionIndex, option, reasoning },
        },
      });
      return response.data;
    } catch (error) {
      dispatch({ type: "SET_ERROR", payload: error.message });
      throw error;
    }
  };

  const getGameState = async () => {
    try {
      const response = await api.get("/state");
      dispatch({ type: "UPDATE_GAME_STATE", payload: response.data });
      return response.data;
    } catch (error) {
      dispatch({ type: "SET_ERROR", payload: error.message });
      throw error;
    }
  };

  const value = {
    ...state,
    startGame,
    newRound,
    evaluateChoice,
    getGameState,
  };

  return <GameContext.Provider value={value}>{children}</GameContext.Provider>;
}

export function useGame() {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error("useGame must be used within a GameProvider");
  }
  return context;
}
