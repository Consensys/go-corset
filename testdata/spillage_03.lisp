(defpurefun (vanishes! x) (== 0 x))

(defcolumns (ST :i16) (A :i16))
(defconstraint spills ()
  (vanishes!
   (* ST A (~ (shift A 3)))))
