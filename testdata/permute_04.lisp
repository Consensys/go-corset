(defcolumns
  (ST :i16@loob@prove)
  (X :i16@loob@prove))
(defpermutation (ST' Y) ((↓ ST) (↑ X)))
(defconstraint first-row (:domain {-1}) (- Y 5))
