(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :binary) (Y :binary) (A :i16))
(defconstraint c1 () (if (== 0 X) (vanishes! A)))
(defconstraint c2 () (if (!= 0 Y) (vanishes! A)))
