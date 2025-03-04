(defpurefun ((vanishes! :@loob) x) x)

(defextern ONE 1)
(defcolumns CT)
(defconstraint c1 ()
  (vanishes!
   (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE)))))
