(defcolumns
  (ST :u16)
  (X :u16))
(defpermutation (ST' Y) ((↓ ST) (↑ X)))
(defconstraint first-row (:domain {-1}) (- Y 5))
