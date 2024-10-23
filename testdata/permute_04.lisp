(defcolumns
  (ST :u16)
  (X :u16))
(permute (ST' Y) (+ST -X))
(defconstraint first-row (:domain {-1}) (- Y 5))
