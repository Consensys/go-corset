;;error:4:43-46:not permitted in pure context
(defpurefun (vanishes! x) (== 0 x))
(defcolumns (X :i16) (Y :i16) (TWO :i16))
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
