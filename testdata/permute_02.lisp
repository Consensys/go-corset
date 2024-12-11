(defcolumns (X :i16@loob@prove))
(defpermutation (Y) ((+ X)))
(defconstraint first-row (:domain {0}) Y)
