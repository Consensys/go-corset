(defcolumns
  (ST :i16@prove)
  (X :i16@prove))
(defpermutation (ST' Y) ((↓ ST) (↑ X)))
(defconstraint first-row (:domain {-1}) (== Y 5))
