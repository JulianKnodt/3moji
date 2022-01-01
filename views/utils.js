import React, { Component, useState, useEffect } from "react";
export const gotoViewTest = (
  stack,
  view,
  currentView,
  setStack,
  setCurrentView
) => {
  setStack([...stack, currentView]);
  setCurrentView(view);
};
