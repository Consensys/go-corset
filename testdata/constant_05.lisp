(defpurefun (vanishes! x) (== 0 x))

(defconst ONE 1)
(defcolumns (CT :i4))
(defconstraint c1 ()
  (vanishes!
   (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE)))))
