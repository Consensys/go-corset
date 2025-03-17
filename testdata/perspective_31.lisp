(defcolumns
  (ST :i16@prove)
  (P :binary@prove))
;;
(defperspective p1 P ((X :i16@prove)))
(defpermutation (ST' Y) ((↓ ST) (↑ p1/X)))
(defconstraint first-row (:domain {-1}) (== Y 5))
