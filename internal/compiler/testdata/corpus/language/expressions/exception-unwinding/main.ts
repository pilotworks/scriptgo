class UnwindTracker {
  trail: string = "";

  stepC(fail: boolean): void {
    this.trail = this.trail + "->stepC";
    try {
      if (fail) {
        throw "range out of bounds";
      }
      this.trail = this.trail + "(ok)";
    } finally {
      this.trail = this.trail + "[finallyC]";
    }
  }

  stepB(fail: boolean): void {
    this.trail = this.trail + "->stepB";
    try {
      this.stepC(fail);
      this.trail = this.trail + "(B_afterC)";
    } catch (err) {
      this.trail = this.trail + "[catchB:" + err + "]";
    } finally {
      this.trail = this.trail + "[finallyB]";
    }
  }

  stepA(fail: boolean): void {
    this.trail = this.trail + "->stepA";
    try {
      this.stepB(fail);
      this.trail = this.trail + "(A_afterB)";
    } finally {
      this.trail = this.trail + "[finallyA]";
    }
  }
}

const tracker1 = new UnwindTracker();
tracker1.stepA(false);
console.log(tracker1.trail);

const tracker2 = new UnwindTracker();
tracker2.stepA(true);
console.log(tracker2.trail);
