(defcolumns (A :i16) (B :i16))
(defconstraint eq (:domain {0}) (== A B))
(defproperty lem (:domain {0}) (== A B))
