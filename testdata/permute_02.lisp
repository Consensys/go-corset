(defcolumns (X :u16))
(defpermutation (Y) ((+ X)))
(defconstraint first-row (:domain {0}) Y)
