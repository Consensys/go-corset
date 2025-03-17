(defcolumns (X :i16@prove))
(defpermutation (Y) ((+ X)))
(defconstraint first-row (:domain {0}) (== 0 Y))
