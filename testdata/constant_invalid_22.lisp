;;error:9:23-26:not permitted in const context
;;error:9:53-56:not permitted in const context
(defpurefun (vanishes! x) (== 0 x))

(defconst (ONE :extern) 1)
(defcolumns (CT :i5))
(defconstraint c1 ()
  (vanishes!
   (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE)))))
