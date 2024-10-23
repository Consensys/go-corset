(defcolumns (X :u16))
(permute (Y) (+X))
(defconstraint first-row (:domain {0}) Y)
