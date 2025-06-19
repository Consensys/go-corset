(defpurefun (vanishes! x) (== 0 x))

(defcolumns (A :i16) (B :i16))
(defconstraint eq () (== A B))
(defproperty lem (== A B))
